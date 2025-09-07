package liveflux

import (
	"context"
	"encoding/json"
	"html"
	"log"
	"net/http"

	"github.com/spf13/cast"
)

// Form field names used by the handler
const (
	FormComponent = "liveflux_component_type"
	FormID        = "liveflux_component_id"
	FormAction    = "livefuse_action"
)

// Response header names for client-side redirect handling
const (
	RedirectHeader       = "X-Liveflux-Redirect"
	RedirectAfterHeader  = "X-Liveflux-Redirect-After"
)

// Handler is an http.Handler that mounts/handles components and returns HTML.
//
// Usage patterns (client-side):
// - To mount: POST with form field `liveflux_component_type` (alias) (and optional params) -> returns initial HTML
// - To act:   POST with `liveflux_component_type` (alias), `liveflux_component_id`, `livefuse_action` (+ any user fields) -> returns updated HTML
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

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid form"))
		return
	}

	alias := r.FormValue(FormComponent)
	id := r.FormValue(FormID)
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

func (h *Handler) mount(ctx context.Context, w http.ResponseWriter, r *http.Request, alias string) {
	if alias == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("missing component alias"))
		return
	}

	c, err := newByAlias(alias)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	id := NewID()
	c.SetID(id)

	params := map[string]string{}
	for key := range r.Form {
		if key == FormComponent || key == FormID || key == FormAction {
			continue
		}
		params[key] = r.Form.Get(key)
	}

	if err := c.Mount(ctx, params); err != nil {
		log.Printf("liveflux: mount error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("mount error"))
		return
	}

	h.Store.Set(c)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(c.Render(ctx).ToHTML()))
}

func (h *Handler) handle(ctx context.Context, w http.ResponseWriter, r *http.Request, alias, id, action string) {
	// Validate basic inputs
	if !h.validateAliasAndID(w, alias, id) {
		return
	}

	c, ok := h.Store.Get(id)
	if !ok || c == nil {
		h.writeError(w, http.StatusNotFound, "component not found")
		return
	}

	// Optional: ensure the retrieved instance matches requested alias by type registry alias.
	// Skipped for simplicity.

	if action != "" {
		if !h.processAction(ctx, w, c, r) {
			return
		}
		// persist after mutation
		h.Store.Set(c)
	}

	if h.maybeWriteRedirect(w, c) {
		return
	}

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
func (h *Handler) processAction(ctx context.Context, w http.ResponseWriter, c Component, r *http.Request) bool {
	if err := c.Handle(ctx, r.FormValue(FormAction), r.Form); err != nil {
		log.Printf("liveflux: handle error: %v", err)
		h.writeError(w, http.StatusBadRequest, "action error")
		return false
	}
	return true
}

// maybeWriteRedirect sends redirect headers and a fallback HTML body if the component requested a redirect.
// Returns true if a redirect response was written.
func (h *Handler) maybeWriteRedirect(w http.ResponseWriter, c Component) bool {
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

// writeRender renders component HTML.
func (h *Handler) writeRender(ctx context.Context, w http.ResponseWriter, c Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(c.Render(ctx).ToHTML()))
}

// writeError writes a status code with a small text message.
func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
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
