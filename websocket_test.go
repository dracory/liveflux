package liveflux

import (
	"context"
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

func (c *fakeWSComponent) GetAlias() string { return "fake-ws" }
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
	u, _ := url.Parse(serverURL)
	u.Scheme = map[string]string{"http": "ws", "https": "wss"}[u.Scheme]
	return websocket.DefaultDialer.Dial(u.String(), nil)
}

func TestWebSocketHandler_ActionFlow(t *testing.T) {
	// Setup handler with default store
	h := NewWebSocketHandler(nil)

	// Create and store a component instance
	comp := &fakeWSComponent{}
	comp.SetAlias(comp.GetAlias())
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

func (c *nonWSComponent) GetAlias() string                                 { return "non-ws" }
func (c *nonWSComponent) Mount(context.Context, map[string]string) error   { return nil }
func (c *nonWSComponent) Handle(context.Context, string, url.Values) error { return nil }
func (c *nonWSComponent) Render(context.Context) hb.TagInterface           { return hb.Div() }

func TestWebSocketHandler_ComponentDoesNotSupportWS(t *testing.T) {
	h := NewWebSocketHandler(nil)

	comp := &nonWSComponent{}
	comp.SetAlias(comp.GetAlias())
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
	comp.SetAlias(comp.GetAlias())
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
