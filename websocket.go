package liveflux

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	// DefaultWebSocketUpgrader is the default upgrader for WebSocket connections.
	DefaultWebSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Allow all origins - in production, you should set a proper origin policy
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// WebSocketMessage represents a message sent over WebSocket.
type WebSocketMessage struct {
	Type        string          `json:"type"`             // message type (e.g., "action", "update", "error")
	ComponentID string          `json:"componentID"`      // ID of the target component
	Action      string          `json:"action,omitempty"` // action name (for action messages)
	Data        json.RawMessage `json:"data,omitempty"`   // message payload
}

// WebSocketComponent is the interface that components can implement to handle WebSocket messages.
type WebSocketComponent interface {
	Component
	// HandleWS handles WebSocket messages for this component.
	// It should return a response message or an error.
	HandleWS(ctx context.Context, message *WebSocketMessage) (interface{}, error)
}

// WebSocketHandler handles WebSocket connections for LiveFlux.
type WebSocketHandler struct {
	*Handler
	upgrader     websocket.Upgrader
	mu           sync.RWMutex
	clients      map[string]map[*websocket.Conn]bool // componentID -> connections
	constructors map[string]func() Component         // alias -> constructor
}

// NewWebSocketHandler creates a new WebSocketHandler.
func NewWebSocketHandler(store Store) *WebSocketHandler {
	return &WebSocketHandler{
		Handler:      NewHandler(store),
		upgrader:     DefaultWebSocketUpgrader,
		clients:      make(map[string]map[*websocket.Conn]bool),
		constructors: make(map[string]func() Component),
	}
}

// Handle registers a component constructor for the given alias.
func (h *WebSocketHandler) Handle(alias string, constructor func() Component) {
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
			log.Printf("WebSocket error on first read: %v", err)
		}
		return
	}
	if firstMsg.ComponentID == "" {
		h.sendError(conn, "missing component ID", http.StatusBadRequest)
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
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Guard against accidental mismatched component IDs from the client
		if msg.ComponentID == "" {
			msg.ComponentID = componentID
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
