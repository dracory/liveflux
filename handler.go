package liveflux

import (
	"context"
	"encoding/json"
	"html"
	"log"
	"net/http"
	"sync"

	"github.com/spf13/cast"
)

// Form field names used by the handler
const (
	FormComponent   = "liveflux_component_type"
	FormComponentID = "liveflux_component_id"
	FormAction      = "liveflux_action"
)

// Response header names for client-side redirect handling and events
const (
	RedirectHeader      = "X-Liveflux-Redirect"
	RedirectAfterHeader = "X-Liveflux-Redirect-After"
	EventsHeader        = "X-Liveflux-Events"
)

// Handler is an http.Handler that mounts/handles components and returns HTML.
//
// Usage patterns (client-side):
// - To mount: POST with form field `liveflux_component_type` (alias) (and optional params) -> returns initial HTML
// - To act:   POST with `liveflux_component_type` (alias), `liveflux_component_id`, `liveflux_action` (+ any user fields) -> returns updated HTML
//
// State is stored via the configured Store (default: in-memory). For production,
// wire a session-backed implementation.
type Handler struct {
	Store Store
}

// NewHandler creates a Handler using the provided store. If store is nil, StoreDefault is used.
func NewHandler(store Store) *Handler {
	if store == nil {
		store = StoreDefault
	}
	return &Handler{Store: store}
}

// NewHandlerWS returns a handler that supports both WebSocket upgrades and regular HTTP POST/GET.
// It is a convenience wrapper around NewWebSocketHandler(store) so developers don't have to
// think about which transport is being used.
func NewHandlerWS(store Store) http.Handler {
	return NewWebSocketHandler(store)
}

// NewHandlerEx returns a handler depending on the enableWebSocket flag:
//   - If enableWebSocket is true, it returns a handler that can upgrade to WebSocket and also
//     handle regular HTTP (POST/GET) requests.
//   - If false, it returns the standard HTTP-only handler.
//
// This allows simple usage like: mux.Handle("/liveflux", liveflux.NewHandlerEx(nil, true))
func NewHandlerEx(store Store, enableWebSocket bool) http.Handler {
	if enableWebSocket {
		return NewWebSocketHandler(store)
	}
	return NewHandler(store)
}

// ServeHTTP handles Liveflux HTTP traffic.
//   - GET requests return the bundled client script so the runtime can be loaded via URL.
//   - POST requests mount components (when no ID is provided) or dispatch actions to existing instances.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodGet {
		h.writeClientScript(w)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid form"))
		return
	}

	alias := r.FormValue(FormComponent)
	id := r.FormValue(FormComponentID)
	action := r.FormValue(FormAction)

	ctx := r.Context()

	// Mount new component if no ID present
	if id == "" {
		h.mount(ctx, w, r, alias)
		return
	}

	// Otherwise, handle action (or just re-render if no action)
	h.handle(ctx, w, r, alias, id, action)
}

// mount creates a new component instance and mounts it.
func (h *Handler) mount(ctx context.Context, w http.ResponseWriter, r *http.Request, alias string) {
	// Validate alias
	if alias == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("missing component alias"))
		return
	}

	// Create new component instance
	c, err := newByAlias(alias)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// Generate and set ID
	id := NewID()
	c.SetID(id)

	// Extract parameters
	params := map[string]string{}
	for key := range r.Form {
		// Skip canonical field names
		if key == FormComponent || key == FormComponentID || key == FormAction {
			continue
		}
		params[key] = r.Form.Get(key)
	}

	// Mount the component
	if err := c.Mount(ctx, params); err != nil {
		// Log error to console
		log.Printf("liveflux: mount error: %v", err)

		// Send 500 error to client with generic error message
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("mount error"))
		return
	}

	h.Store.Set(c)

	// Check if component supports events
	if ea, ok := c.(EventAware); ok {
		dispatcher := ea.GetEventDispatcher()
		if dispatcher != nil && dispatcher.HasEvents() {
			// Send events as a header
			w.Header().Set(EventsHeader, dispatcher.EventsJSON())
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(c.Render(ctx).ToHTML()))
}

func (h *Handler) handle(ctx context.Context, w http.ResponseWriter, r *http.Request, alias, id, action string) {
	log.Printf("[Liveflux Handler] handle: alias=%s, id=%s, action=%s", alias, id, action)

	// Validate basic inputs
	if !h.validateAliasAndID(w, alias, id) {
		return
	}

	// Acquire per-component lock to prevent concurrent modifications
	// This is critical when multiple requests target the same component ID
	var componentLock *sync.Mutex
	if memStore, ok := h.Store.(*MemoryStore); ok {
		componentLock = memStore.LockComponent(id)
		defer memStore.UnlockComponent(componentLock)
	}

	// Retrieve component from store
	c, ok := h.Store.Get(id)
	if !ok || c == nil {
		h.writeError(w, http.StatusNotFound, "component not found")
		return
	}

	// Optional: ensure the retrieved instance matches requested alias by type registry alias.
	// Skipped for simplicity.

	// Process action if present
	if action != "" {
		if !h.processAction(ctx, w, c, r) {
			return
		}
		// persist after mutation
		h.Store.Set(c)
	}

	// Handle redirect if requested
	if h.maybeWriteRedirect(w, c) {
		return
	}

	// Render the component
	h.writeRender(ctx, w, c)
}

// validateAliasAndID ensures required params are present. Returns true if OK.
func (h *Handler) validateAliasAndID(w http.ResponseWriter, alias, id string) bool {
	if alias == "" || id == "" {
		h.writeError(w, http.StatusBadRequest, "missing component or id")
		return false
	}
	return true
}

// processAction invokes the component's Handle for the given action. Returns true if successful.
func (h *Handler) processAction(ctx context.Context, w http.ResponseWriter, c ComponentInterface, r *http.Request) bool {
	if err := c.Handle(ctx, r.FormValue(FormAction), r.Form); err != nil {
		log.Printf("liveflux: handle error: %v", err)
		h.writeError(w, http.StatusBadRequest, "action error")
		return false
	}
	return true
}

// maybeWriteRedirect sends redirect headers and a fallback HTML body if the component requested a redirect.
// Returns true if a redirect response was written.
func (h *Handler) maybeWriteRedirect(w http.ResponseWriter, c ComponentInterface) bool {
	redir, ok := c.(interface{ TakeRedirect() string })
	if !ok {
		return false
	}
	url := redir.TakeRedirect()
	if url == "" {
		return false
	}

	// Determine delay once (TakeRedirectDelaySeconds resets it)
	delay := 0
	if rdelay, ok2 := c.(interface{ TakeRedirectDelaySeconds() int }); ok2 {
		delay = rdelay.TakeRedirectDelaySeconds()
	}

	w.Header().Set(RedirectHeader, url)
	if delay > 0 {
		w.Header().Set(RedirectAfterHeader, cast.ToString(delay))
	}

	// Fallback HTML body: <script> redirect (with delay) and <noscript> meta refresh
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(buildRedirectFallbackHTML(url, delay)))
	return true
}

// writeRender renders component HTML and sends any queued events.
func (h *Handler) writeRender(ctx context.Context, w http.ResponseWriter, c ComponentInterface) {
	// Check if component supports events
	if ea, ok := c.(EventAware); ok {
		dispatcher := ea.GetEventDispatcher()
		if dispatcher != nil && dispatcher.HasEvents() {
			eventsJSON := dispatcher.EventsJSON()
			log.Printf("[Liveflux Events] Sending events in response: %s", eventsJSON)
			// Send events as a header
			w.Header().Set(EventsHeader, eventsJSON)
		} else {
			log.Printf("[Liveflux Events] No events to send for component %s", c.GetAlias())
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(c.Render(ctx).ToHTML()))
}

// writeError writes a status code with a small text message.
func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}

func (h *Handler) writeClientScript(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = w.Write([]byte(JS()))
}

// buildRedirectFallbackHTML returns the script + noscript fallback HTML document for a redirect.
func buildRedirectFallbackHTML(url string, delaySeconds int) string {
	urlJSON, _ := json.Marshal(url) // safe JS string literal
	urlEsc := html.EscapeString(url)
	return "<!doctype html><html><head><meta charset=\"utf-8\">" +
		"<noscript><meta http-equiv=\"refresh\" content=\"" + cast.ToString(delaySeconds) + ";url=" + urlEsc + "\"></noscript>" +
		"<script>(function(){var url=" + string(urlJSON) + ";var ms=" + cast.ToString(delaySeconds*1000) + ";var go=function(){window.location.replace(url);};if(ms>0)setTimeout(go,ms);else go();})();</script>" +
		"</head><body></body></html>"
}
