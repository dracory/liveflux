# WebSocket Integration

Liveflux supports real-time updates over WebSockets with graceful fallback to HTTP. The WebSocket layer reuses component state stored in the same `Store` used by the HTTP handler.

## When to Use WebSockets

- High-frequency updates (e.g., dashboards, notifications, collaborative UIs)
- Bidirectional messaging (components proactively sending updates to the client)
- Reduced latency compared to repeated HTTP polling

## Enabling WebSockets

Use `liveflux.NewHandlerWS(store)` or `liveflux.NewWebSocketHandler(store, opts...)` in your router:

```go
mux.Handle("/liveflux", liveflux.NewHandlerWS(nil))
```

On the client, enable WebSockets when rendering the script:

```go
liveflux.Script(liveflux.ClientOptions{
    UseWebSocket: true,
    WebSocketURL: "/liveflux", // optional; defaults to Endpoint
})
```

The client will auto-upgrade mount requests to WebSocket actions after the initial HTTP mount completes.

## WebSocket Handler (`websocket.go`)

`NewWebSocketHandler` extends the base HTTP handler and adds:

- Connection management (`clients` map keyed by component ID)
- Optional CSRF, TLS enforcement, rate limiting, and message validation
- Broadcasting helpers for pushing updates

### Options

- `WithWebSocketAllowedOrigins(origins...)`
- `WithWebSocketCSRFCheck(func(*http.Request) error)`
- `WithWebSocketRequireTLS(true)`
- `WithWebSocketRateLimit(max, window)`
- `WithWebSocketMessageValidator(func(*WebSocketMessage) error)`

## Component Support

Components using WebSockets should implement the `WebSocketComponent` interface:

```go
type WebSocketComponent interface {
    liveflux.Component
    HandleWS(ctx context.Context, message *liveflux.WebSocketMessage) (interface{}, error)
}
```

Messages are encoded as JSON with fields:

```go
type WebSocketMessage struct {
    Type        string          `json:"type"`
    ComponentID string          `json:"componentID"`
    Action      string          `json:"action,omitempty"`
    Data        json.RawMessage `json:"data,omitempty"`
}
```

`HandleWS` returns a response payload that the server writes back as JSON.

## Broadcasts

Use `WebSocketHandler.Broadcast(componentID, payload)` to push updates to all clients watching a component ID. This is useful for server-driven notifications or collaborative edits:

```go
err := wsHandler.Broadcast(componentID, struct {
    Type string `json:"type"`
    Data any    `json:"data"`
}{
    Type: "update",
    Data: state,
})
```

## Fallback Behavior

When the client cannot establish a WebSocket connection (e.g., due to network/firewall restrictions), the handler falls back to standard HTTP form posts. Components should treat actions identically regardless of transport.

## Security Considerations

- Restrict allowed origins using `WithWebSocketAllowedOrigins` or the default upgrader.
- Enforce CSRF tokens via `WithWebSocketCSRFCheck` and custom headers configured in `ClientOptions`.
- Require TLS in production for confidentiality (`WithWebSocketRequireTLS(true)`).
- Use `WithWebSocketMessageValidator` to sanitize incoming messages before they reach components.

## Example Flow

1. Client mounts component over HTTP.
2. Client opens WebSocket connection to `/liveflux` and sends an initial `WebSocketMessage` containing the component ID.
3. Server associates the connection with the component ID.
4. Subsequent messages from the client call `HandleWS`; responses are sent via `WriteJSON`.
5. Server can push updates to connected clients using `Broadcast`.

Refer to `examples/websocket/` for a working demonstration with two counter instances synchronized via WebSockets.
