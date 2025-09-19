# Liveflux WebSocket Counter Example

A minimal, runnable Liveflux example demonstrating WebSocket-based interactions.

## Run

```bash
# from the liveflux repo root
go run ./examples/websocket
```

Then open http://localhost:8080

## Screenshot

<!-- Optional: add a screenshot image next to this README -->
<!-- ![WebSocket Counter — Two instances](./screenshot.png) -->

## What it does
- Renders two server-side `WebSocketCounter` components using `liveflux.SSR`
- Mounts the client runtime via `liveflux.Script()`
- Initializes a WebSocket client via the core client served at `/static/websocket.js`
- Sends actions over a WebSocket connection to the `/liveflux` endpoint
- Receives server-pushed updates and swaps the component HTML in place

## Endpoints
- `GET /` — serves the demo page containing two counter instances
- `WS /liveflux` — WebSocket endpoint handled by `liveflux.NewWebSocketHandler(nil)`
- `GET /static/websocket.js` — serves the core WebSocket client from `js/websocket.js`
- `GET /static/*` — other static assets

## Files
- `main.go` — wires the HTTP mux, WebSocket handler, SSR, and static assets
- `websocket_counter.go` — the component implementation (implements both `Handle` and `HandleWS`)
- `../../js/websocket.js` — Core WebSocket client that upgrades form/button actions to WS

## Notes
- The component wraps its markup with `c.Root(content)` (provided by `liveflux.Base`) to include the standard Liveflux root and required hidden fields (`component`, `id`).
- The wrapper element includes `data-flux-ws` and `data-flux-ws-url="/liveflux"` so the JS client connects to the correct endpoint.
- The `Handle` method uses `url.Values` (not `map[string]string`) to conform to the `Component` interface.
