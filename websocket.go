package liveflux

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// DefaultWebSocketUpgrader is the default upgrader for WebSocket connections.
	DefaultWebSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Allow all origins - in production, you should set a proper origin policy
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// WithWebSocketCSRFCheck configures a custom CSRF validation function that is
// executed before upgrading the HTTP connection to a WebSocket. Returning a
// non-nil error will reject the upgrade with HTTP 403.
func WithWebSocketCSRFCheck(check func(*http.Request) error) WebSocketOption {
	return func(opts *websocketOptions) {
		opts.csrfCheck = check
	}
}

// WithWebSocketRequireTLS enforces that WebSocket upgrades occur over HTTPS/WSS.
// When enabled, plaintext HTTP upgrade attempts are rejected with HTTP 403.
func WithWebSocketRequireTLS(require bool) WebSocketOption {
	return func(opts *websocketOptions) {
		opts.requireTLS = require
	}
}

// WithWebSocketRateLimit configures a simple per-client connection rate limit
// for WebSocket upgrades. A max value <= 0 disables limiting. The window
// defines the period for counting attempts; if zero, one minute is used.
func WithWebSocketRateLimit(max int, window time.Duration) WebSocketOption {
	return func(opts *websocketOptions) {
		opts.rateLimitMax = max
		opts.rateLimitWindow = window
	}
}

// WithWebSocketMessageValidator sets a callback invoked for every WebSocket
// message before it is dispatched to the component. Returning an error sends a
// 400 response/error frame to the client and skips processing.
func WithWebSocketMessageValidator(validator func(*WebSocketMessage) error) WebSocketOption {
	return func(opts *websocketOptions) {
		opts.messageValidator = validator
	}
}

type websocketOptions struct {
	allowedOrigins   []string
	csrfCheck        func(*http.Request) error
	requireTLS       bool
	rateLimitMax     int
	rateLimitWindow  time.Duration
	messageValidator func(*WebSocketMessage) error
}

// WebSocketOption configures optional behaviour for the WebSocket handler.
type WebSocketOption func(*websocketOptions)

func defaultWebSocketOptions() websocketOptions {
	return websocketOptions{}
}

// WithWebSocketAllowedOrigins restricts allowed WebSocket upgrade origins to the
// provided list. Comparisons are case-insensitive and expect fully qualified
// origin strings (scheme://host[:port]). When no origins are specified the
// handler falls back to the default upgrader behaviour (allow all).
func WithWebSocketAllowedOrigins(origins ...string) WebSocketOption {
	trimmed := make([]string, 0, len(origins))
	for _, origin := range origins {
		o := strings.TrimSpace(origin)
		if o == "" {
			continue
		}
		trimmed = append(trimmed, o)
	}
	return func(opts *websocketOptions) {
		opts.allowedOrigins = trimmed
	}
}

// WebSocketMessage represents a message sent over WebSocket.
type WebSocketMessage struct {
	Type        string          `json:"type"`             // message type (e.g., "action", "update", "error")
	ComponentID string          `json:"componentID"`      // ID of the target component
	Action      string          `json:"action,omitempty"` // action name (for action messages)
	Data        json.RawMessage `json:"data,omitempty"`   // message payload
}

// WebSocketComponent is the interface that components can implement to handle WebSocket messages.
type WebSocketComponent interface {
	ComponentInterface
	// HandleWS handles WebSocket messages for this component.
	// It should return a response message or an error.
	HandleWS(ctx context.Context, message *WebSocketMessage) (any, error)
}

// WebSocketHandler handles WebSocket connections for LiveFlux.
type WebSocketHandler struct {
	*Handler
	upgrader         websocket.Upgrader
	mu               sync.RWMutex
	clients          map[string]map[*websocket.Conn]bool  // componentID -> connections
	constructors     map[string]func() ComponentInterface // alias -> constructor
	allowedOrigins   []string
	csrfCheck        func(*http.Request) error
	requireTLS       bool
	rateLimiter      *wsRateLimiter
	messageValidator func(*WebSocketMessage) error
}

// NewWebSocketHandler creates a new WebSocketHandler.
func NewWebSocketHandler(store Store, optFns ...WebSocketOption) *WebSocketHandler {
	options := defaultWebSocketOptions()
	for _, fn := range optFns {
		if fn != nil {
			fn(&options)
		}
	}

	h := &WebSocketHandler{
		Handler:          NewHandler(store),
		upgrader:         DefaultWebSocketUpgrader,
		clients:          make(map[string]map[*websocket.Conn]bool),
		constructors:     make(map[string]func() ComponentInterface),
		allowedOrigins:   append([]string(nil), options.allowedOrigins...),
		csrfCheck:        options.csrfCheck,
		requireTLS:       options.requireTLS,
		rateLimiter:      newWSRateLimiter(options.rateLimitMax, options.rateLimitWindow),
		messageValidator: options.messageValidator,
	}

	defaultCheck := DefaultWebSocketUpgrader.CheckOrigin
	h.upgrader.CheckOrigin = func(r *http.Request) bool {
		return h.checkOrigin(r, defaultCheck)
	}

	return h
}

// checkOrigin validates the upgrade origin against the allow-list, falling back
// to the upstream upgrader behaviour when no list is configured.
func (h *WebSocketHandler) checkOrigin(r *http.Request, defaultCheck func(*http.Request) bool) bool {
	if len(h.allowedOrigins) == 0 {
		return defaultCheck(r)
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}

	for _, allowed := range h.allowedOrigins {
		if strings.EqualFold(origin, allowed) {
			return true
		}
	}

	return false
}

// Handle registers a component constructor for the given alias.
func (h *WebSocketHandler) Handle(alias string, constructor func() ComponentInterface) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.constructors[alias] = constructor
}

// ServeHTTP implements http.Handler.
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if this is a WebSocket upgrade request
	if websocket.IsWebSocketUpgrade(r) {
		h.handleWebSocket(w, r)
		return
	}

	// Fall back to regular HTTP handler
	h.Handler.ServeHTTP(w, r)
}

// handleWebSocket handles a WebSocket connection.
func (h *WebSocketHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if h.requireTLS && r.TLS == nil {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if h.rateLimiter != nil {
		ip := clientIP(r)
		if !h.rateLimiter.Allow(ip) {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
	}

	if h.csrfCheck != nil {
		if err := h.csrfCheck(r); err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// The upgrader has already written an error response
		return
	}
	defer conn.Close()

	// Set up a context for this connection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Read the initial message to learn the componentID and process it
	var firstMsg WebSocketMessage
	if err := conn.ReadJSON(&firstMsg); err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			fmt.Printf("WebSocket error on first read: %v\n", err)
		}
		return
	}
	if firstMsg.ComponentID == "" {
		h.sendError(conn, "missing component ID", http.StatusBadRequest)
		return
	}

	if !h.validateMessage(conn, &firstMsg) {
		return
	}

	// Register and ensure cleanup
	componentID := firstMsg.ComponentID
	h.registerConnection(componentID, conn)
	defer h.unregisterConnection(componentID, conn)

	// Handle the initial message
	h.handleMessage(ctx, conn, &firstMsg)

	// Continue handling subsequent messages
	for {
		var msg WebSocketMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// Guard against accidental mismatched component IDs from the client
		if msg.ComponentID == "" {
			msg.ComponentID = componentID
		}

		if !h.validateMessage(conn, &msg) {
			continue
		}

		// Handle the message in a goroutine
		go h.handleMessage(ctx, conn, &msg)
	}
}

// handleMessage processes a WebSocket message.
func (h *WebSocketHandler) handleMessage(ctx context.Context, conn *websocket.Conn, msg *WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get the component from the store
	c, found := h.Store.Get(msg.ComponentID)
	if !found {
		h.sendError(conn, "component not found", http.StatusNotFound)
		return
	}

	// Check if the component supports WebSocket
	wsComp, ok := c.(WebSocketComponent)
	if !ok {
		h.sendError(conn, "component does not support WebSocket", http.StatusBadRequest)
		return
	}

	// Handle the message
	resp, wsErr := wsComp.HandleWS(ctx, msg)
	if wsErr != nil {
		h.sendError(conn, wsErr.Error(), http.StatusInternalServerError)
		return
	}

	// Send the response
	if resp != nil {
		if err := conn.WriteJSON(resp); err != nil {
			// Connection is probably closed
			return
		}
	}
}

// registerConnection registers a WebSocket connection for a component.
func (h *WebSocketHandler) registerConnection(componentID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[componentID] == nil {
		h.clients[componentID] = make(map[*websocket.Conn]bool)
	}
	h.clients[componentID][conn] = true
}

// unregisterConnection removes a WebSocket connection.
func (h *WebSocketHandler) unregisterConnection(componentID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if connections, ok := h.clients[componentID]; ok {
		delete(connections, conn)
		if len(connections) == 0 {
			delete(h.clients, componentID)
		}
	}
}

// Broadcast sends a message to all connected clients for a component.
func (h *WebSocketHandler) Broadcast(componentID string, message interface{}) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connections, ok := h.clients[componentID]
	if !ok {
		return nil // No clients connected
	}

	var errs []error
	for conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		// Return the first error
		return errs[0]
	}
	return nil
}

// sendError sends an error message to the client.
func (h *WebSocketHandler) sendError(conn *websocket.Conn, message string, code int) {
	errMsg := struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}{
		Type:    "error",
		Message: message,
		Code:    code,
	}

	_ = conn.WriteJSON(errMsg)
}

func (h *WebSocketHandler) validateMessage(conn *websocket.Conn, msg *WebSocketMessage) bool {
	if h.messageValidator == nil {
		return true
	}
	if err := h.messageValidator(msg); err != nil {
		h.sendError(conn, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

type wsRateLimiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	entries map[string]*wsRateEntry
}

type wsRateEntry struct {
	count       int
	windowStart time.Time
}

func newWSRateLimiter(max int, window time.Duration) *wsRateLimiter {
	if max <= 0 {
		return nil
	}
	if window <= 0 {
		window = time.Minute
	}
	return &wsRateLimiter{
		max:     max,
		window:  window,
		entries: make(map[string]*wsRateEntry),
	}
}

func (rl *wsRateLimiter) Allow(ip string) bool {
	if rl == nil {
		return true
	}
	if ip == "" {
		ip = "unknown"
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry := rl.entries[ip]
	if entry == nil || now.Sub(entry.windowStart) >= rl.window {
		rl.entries[ip] = &wsRateEntry{count: 1, windowStart: now}
		return true
	}

	if entry.count >= rl.max {
		return false
	}

	entry.count++
	return true
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
