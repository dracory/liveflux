package liveflux

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/dracory/hb"
	"github.com/gorilla/websocket"
)

// fakeWSComponent is a minimal component implementing WebSocketComponent
// used for testing the websocket handler.
type fakeWSComponent struct {
	Base
	Count int
}

func TestWebSocketHandler_MessageValidatorPasses(t *testing.T) {
	validator := func(msg *WebSocketMessage) error {
		if msg.Action == "block" {
			return errors.New("blocked")
		}
		return nil
	}

	store := NewMemoryStore()
	h := NewWebSocketHandler(store, WithWebSocketMessageValidator(validator))

	comp := &fakeWSComponent{}
	comp.SetKind(comp.GetKind())
	comp.SetID(NewID())
	if err := comp.Mount(context.Background(), map[string]string{}); err != nil {
		t.Fatalf("mount error: %v", err)
	}
	store.Set(comp)

	ts := httptest.NewServer(h)
	defer ts.Close()

	conn, resp, err := dialWS(t, ts.URL)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("unexpected HTTP response on upgrade: %d", resp.StatusCode)
	}
	defer func() {
		_ = conn.Close()
	}()

	msg := WebSocketMessage{Type: "action", ComponentID: comp.GetID(), Action: "inc"}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestWebSocketHandler_MessageValidatorBlocks(t *testing.T) {
	validator := func(msg *WebSocketMessage) error {
		if msg.Action == "block" {
			return errors.New("blocked")
		}
		return nil
	}

	h := NewWebSocketHandler(nil, WithWebSocketMessageValidator(validator))
	ts := httptest.NewServer(h)
	defer ts.Close()

	conn, resp, err := dialWS(t, ts.URL)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("unexpected HTTP response on upgrade: %d", resp.StatusCode)
	}
	defer conn.Close()

	err = conn.WriteJSON(WebSocketMessage{Type: "action", ComponentID: "comp", Action: "block"})
	if err != nil {
		t.Fatalf("write block action: %v", err)
	}

	var failure struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	if err := conn.ReadJSON(&failure); err != nil {
		t.Fatalf("read error: %v", err)
	}
	if failure.Type != "error" || failure.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 error frame, got %#v", failure)
	}
}

func TestWebSocketHandler_RequireTLSRejectsPlain(t *testing.T) {
	h := NewWebSocketHandler(nil, WithWebSocketRequireTLS(true))

	ts := httptest.NewServer(h)
	defer ts.Close()

	_, resp, err := dialWS(t, ts.URL)
	if err == nil {
		t.Fatal("expected dial error for non-TLS upgrade")
	}
	if resp == nil {
		t.Fatal("expected HTTP response on failed upgrade")
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestWebSocketHandler_RequireTLSPermitsSecure(t *testing.T) {
	store := NewMemoryStore()
	h := NewWebSocketHandler(store, WithWebSocketRequireTLS(true))

	comp := &fakeWSComponent{}
	comp.SetKind(comp.GetKind())
	comp.SetID(NewID())
	if err := comp.Mount(context.Background(), map[string]string{}); err != nil {
		t.Fatalf("mount error: %v", err)
	}
	store.Set(comp)

	ts := httptest.NewTLSServer(h)
	defer ts.Close()

	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	conn, _, err := dialWSWithOptions(t, ts.URL, nil, &dialer)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	msg := WebSocketMessage{Type: "action", ComponentID: comp.GetID(), Action: "inc"}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestWebSocketHandler_RateLimitAllowsWithinThreshold(t *testing.T) {
	ratelimited := NewWebSocketHandler(nil, WithWebSocketRateLimit(2, time.Second))
	ts := httptest.NewServer(ratelimited)
	defer ts.Close()

	for i := 0; i < 2; i++ {
		conn, resp, err := dialWS(t, ts.URL)
		if err != nil {
			t.Fatalf("unexpected dial error on attempt %d: %v", i, err)
		}
		if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
			t.Fatalf("unexpected HTTP response on upgrade: %d", resp.StatusCode)
		}
		_ = conn.Close()
	}
}

func TestWebSocketHandler_RateLimitBlocksExcess(t *testing.T) {
	ratelimited := NewWebSocketHandler(nil, WithWebSocketRateLimit(1, time.Minute))
	ts := httptest.NewServer(ratelimited)
	defer ts.Close()

	conn, resp, err := dialWS(t, ts.URL)
	if err != nil {
		t.Fatalf("unexpected dial error for first attempt: %v", err)
	}
	if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("unexpected HTTP response on upgrade: %d", resp.StatusCode)
	}
	conn.Close()

	_, resp2, err := dialWS(t, ts.URL)
	if err == nil {
		t.Fatal("expected rate limit dial error on second attempt")
	}
	if resp2 == nil {
		t.Fatal("expected HTTP response on blocked upgrade")
	}
	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp2.StatusCode)
	}
}

func TestWebSocketHandler_CSRFCheckPasses(t *testing.T) {
	check := func(r *http.Request) error {
		if r.Header.Get("X-CSRF") != "token" {
			return errors.New("missing token")
		}
		return nil
	}

	store := NewMemoryStore()
	h := NewWebSocketHandler(store, WithWebSocketCSRFCheck(check))

	comp := &fakeWSComponent{}
	comp.SetKind(comp.GetKind())
	comp.SetID(NewID())
	if err := comp.Mount(context.Background(), map[string]string{}); err != nil {
		t.Fatalf("mount error: %v", err)
	}
	store.Set(comp)

	ts := httptest.NewServer(h)
	defer ts.Close()

	headers := http.Header{}
	headers.Set("X-CSRF", "token")
	conn, _, err := dialWSWithHeader(t, ts.URL, headers)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	msg := WebSocketMessage{Type: "action", ComponentID: comp.GetID(), Action: "inc"}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestWebSocketHandler_CSRFCheckBlocks(t *testing.T) {
	check := func(*http.Request) error { return errors.New("denied") }
	h := NewWebSocketHandler(nil, WithWebSocketCSRFCheck(check))

	ts := httptest.NewServer(h)
	defer ts.Close()

	_, resp, err := dialWSWithHeader(t, ts.URL, nil)
	if err == nil {
		t.Fatal("expected dial error due to CSRF")
	}
	if resp == nil {
		t.Fatal("expected HTTP response on failed upgrade")
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func (c *fakeWSComponent) GetKind() string { return "fake-ws" }
func (c *fakeWSComponent) Mount(ctx context.Context, params map[string]string) error {
	c.Count = 0
	return nil
}
func (c *fakeWSComponent) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "inc":
		c.Count++
	}
	return nil
}
func (c *fakeWSComponent) Render(ctx context.Context) hb.TagInterface { return hb.Div() }

func (c *fakeWSComponent) HandleWS(ctx context.Context, msg *WebSocketMessage) (interface{}, error) {
	if msg.Type == "action" && msg.Action == "inc" {
		c.Count++
		return map[string]any{
			"type":        "update",
			"componentID": c.GetID(),
			"data":        map[string]any{"count": c.Count},
		}, nil
	}
	return nil, nil
}

func dialWS(t *testing.T, serverURL string) (*websocket.Conn, *http.Response, error) {
	return dialWSWithOptions(t, serverURL, nil, nil)
}

func dialWSWithHeader(t *testing.T, serverURL string, header http.Header) (*websocket.Conn, *http.Response, error) {
	return dialWSWithOptions(t, serverURL, header, nil)
}

func dialWSWithOptions(t *testing.T, serverURL string, header http.Header, dialer *websocket.Dialer) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	u, err := url.Parse(serverURL)
	if err != nil || u == nil {
		return nil, nil, err
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	if dialer == nil {
		dialer = websocket.DefaultDialer
	}

	return dialer.Dial(u.String(), header)
}

func TestWebSocketHandler_ActionFlow(t *testing.T) {
	// Setup handler with default store
	h := NewWebSocketHandler(nil)

	// Create and store a component instance
	comp := &fakeWSComponent{}
	comp.SetKind(comp.GetKind())
	comp.SetID(NewID())
	if err := comp.Mount(context.Background(), map[string]string{}); err != nil {
		t.Fatalf("mount error: %v", err)
	}
	StoreDefault.Set(comp)

	// Serve handler
	ts := httptest.NewServer(h)
	defer ts.Close()

	// Connect WS
	conn, _, err := dialWS(t, ts.URL)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	// Send initial message with componentID and an action
	init := WebSocketMessage{Type: "action", ComponentID: comp.GetID(), Action: "inc"}
	if err := conn.WriteJSON(init); err != nil {
		t.Fatalf("write init: %v", err)
	}

	// Expect an update message
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var resp map[string]any
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read resp: %v", err)
	}

	if resp["type"] != "update" {
		t.Fatalf("expected update, got %v", resp["type"])
	}
	data, _ := resp["data"].(map[string]any)
	if data == nil || int(data["count"].(float64)) != 1 {
		t.Fatalf("expected count=1, got %#v", resp)
	}
}

func TestWebSocketHandler_MissingComponentID(t *testing.T) {
	h := NewWebSocketHandler(nil)
	ts := httptest.NewServer(h)
	defer ts.Close()

	conn, _, err := dialWS(t, ts.URL)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	// Send message without componentID
	msg := WebSocketMessage{Type: "action", Action: "inc"}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Expect an error message
	var resp struct {
		Type, Message string
		Code          int
	}
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read: %v", err)
	}
	if resp.Type != "error" || resp.Code != 400 {
		t.Fatalf("expected error 400, got %#v", resp)
	}
}

// nonWSComponent does not implement WebSocketComponent
type nonWSComponent struct{ Base }

func (c *nonWSComponent) GetKind() string                                  { return "non-ws" }
func (c *nonWSComponent) Mount(context.Context, map[string]string) error   { return nil }
func (c *nonWSComponent) Handle(context.Context, string, url.Values) error { return nil }
func (c *nonWSComponent) Render(context.Context) hb.TagInterface           { return hb.Div() }

func TestWebSocketHandler_ComponentDoesNotSupportWS(t *testing.T) {
	h := NewWebSocketHandler(nil)

	comp := &nonWSComponent{}
	comp.SetKind(comp.GetKind())
	comp.SetID(NewID())
	StoreDefault.Set(comp)

	ts := httptest.NewServer(h)
	defer ts.Close()

	conn, _, err := dialWS(t, ts.URL)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	msg := WebSocketMessage{Type: "action", ComponentID: comp.GetID(), Action: "x"}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write: %v", err)
	}

	var resp struct {
		Type, Message string
		Code          int
	}
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read: %v", err)
	}
	if resp.Type != "error" || resp.Code != 400 {
		t.Fatalf("expected error 400, got %#v", resp)
	}
}

func TestWebSocketHandler_UnregisterOnClose(t *testing.T) {
	h := NewWebSocketHandler(nil)

	comp := &fakeWSComponent{}
	comp.SetKind(comp.GetKind())
	comp.SetID(NewID())
	StoreDefault.Set(comp)

	ts := httptest.NewServer(h)
	defer ts.Close()

	conn, _, err := dialWS(t, ts.URL)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}

	// send initial message with componentID
	_ = conn.WriteJSON(WebSocketMessage{Type: "action", ComponentID: comp.GetID(), Action: "inc"})
	conn.Close()

	// Allow cleanup to happen
	time.Sleep(100 * time.Millisecond)

	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.clients[comp.GetID()]) != 0 {
		t.Fatalf("expected no clients after close")
	}
}

func TestWebSocketHandler_AllowedOriginsPermitMatch(t *testing.T) {
	h := NewWebSocketHandler(nil, WithWebSocketAllowedOrigins("https://example.com"))
	req := httptest.NewRequest(http.MethodGet, "http://server/liveflux", nil)
	req.Header.Set("Origin", "https://example.com")

	if !h.upgrader.CheckOrigin(req) {
		t.Fatalf("expected matching origin to be allowed")
	}
}

func TestWebSocketHandler_AllowedOriginsRejectMismatch(t *testing.T) {
	h := NewWebSocketHandler(nil, WithWebSocketAllowedOrigins("https://example.com"))

	req := httptest.NewRequest(http.MethodGet, "http://server/liveflux", nil)
	req.Header.Set("Origin", "https://other.com")
	if h.upgrader.CheckOrigin(req) {
		t.Fatalf("expected mismatched origin to be rejected")
	}

	noOrigin := httptest.NewRequest(http.MethodGet, "http://server/liveflux", nil)
	if h.upgrader.CheckOrigin(noOrigin) {
		t.Fatalf("expected missing origin to be rejected when allow list configured")
	}
}
