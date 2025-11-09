package liveflux

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"sync"

	"github.com/spf13/cast"
)

// Form field names used by the handler
const (
	FormComponentKind = "liveflux_component_kind"
	FormComponentID   = "liveflux_component_id"
	FormAction        = "liveflux_action"
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
// - To mount: POST with form field `liveflux_component_kind` (kind) (and optional params) -> returns initial HTML
// - To act:   POST with `liveflux_component_kind` (kind), `liveflux_component_id`, `liveflux_action` (+ any user fields) -> returns updated HTML
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

	kind := r.FormValue(FormComponentKind)
	id := r.FormValue(FormComponentID)
	action := r.FormValue(FormAction)

	ctx := r.Context()

	// Mount new component if no ID present
	if id == "" {
		h.mount(ctx, w, r, kind)
		return
	}

	// Otherwise, handle action (or just re-render if no action)
	h.handle(ctx, w, r, kind, id, action)
}

// mount creates a new component instance and mounts it.
func (h *Handler) mount(ctx context.Context, w http.ResponseWriter, r *http.Request, kind string) {
	// Validate kind
	if kind == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("missing component kind"))
		return
	}

	// Create new component instance
	c, err := newByKind(kind)
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
		if key == FormComponentKind || key == FormComponentID || key == FormAction {
			continue
		}
		params[key] = r.Form.Get(key)
	}

	// Mount the component
	if err := c.Mount(ctx, params); err != nil {
		// Log error to console
		fmt.Printf("liveflux: mount error: %v\n", err)

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

func (h *Handler) handle(ctx context.Context, w http.ResponseWriter, r *http.Request, kind, id, action string) {
	fmt.Printf("[Liveflux Handler] handle: kind=%s, id=%s, action=%s\n", kind, id, action)

	// Validate basic inputs
	if !h.validateKindAndID(w, kind, id) {
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

	// Optional: ensure the retrieved instance matches requested kind by type registry kind.
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
	h.writeRender(ctx, w, r, c)
}

// validateKindAndID ensures required params are present. Returns true if OK.
func (h *Handler) validateKindAndID(w http.ResponseWriter, kind, id string) bool {
	if kind == "" || id == "" {
		h.writeError(w, http.StatusBadRequest, "missing component or id")
		return false
	}
	return true
}

// processAction invokes the component's Handle for the given action. Returns true if successful.
func (h *Handler) processAction(ctx context.Context, w http.ResponseWriter, c ComponentInterface, r *http.Request) bool {
	if err := c.Handle(ctx, r.FormValue(FormAction), r.Form); err != nil {
		fmt.Printf("liveflux: handle error: %v\n", err)
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
// If the component implements TargetRenderer, it will send only the changed fragments
// instead of the full component.
func (h *Handler) writeRender(ctx context.Context, w http.ResponseWriter, r *http.Request, c ComponentInterface) {
	// Check if component supports events
	if ea, ok := c.(EventAware); ok {
		dispatcher := ea.GetEventDispatcher()
		if dispatcher != nil && dispatcher.HasEvents() {
			eventsJSON := dispatcher.EventsJSON()
			fmt.Printf("[Liveflux Events] Sending events in response: %s\n", eventsJSON)
			// Send events as a header
			w.Header().Set(EventsHeader, eventsJSON)
		} else {
			fmt.Printf("[Liveflux Events] No events to send for component %s\n", c.GetKind())
		}
	}

	// Try targeted rendering if component implements TargetRenderer
	if tr, ok := c.(TargetRenderer); ok {
		fragments := tr.RenderTargets(ctx)
		if len(fragments) > 0 {
			// Build template response with fragments only (no fallback)
			// The client will handle fallback if selectors fail
			response := BuildTargetResponse(fragments, "", c)

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(response))
			return
		}
	}

	// Fallback to full render
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
