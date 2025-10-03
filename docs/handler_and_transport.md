# Handler and Transport

Liveflux exposes HTTP handlers that integrate with any Go router. This guide explains how the request pipeline works, how to configure transports, and how to integrate with existing middleware.

## HTTP Handler (`handler.go`)

`liveflux.NewHandler(store Store) *Handler` returns an `http.Handler` that accepts `POST` and `GET` requests. The handler expects the following form fields:

- `liveflux_component_type` (`FormComponent` constant): component alias.
- `liveflux_component_id` (`FormID`): assigned during mount; required for actions.
- `liveflux_action` (`FormAction`): optional action identifier.

### Mount Requests

1. Clients submit forms without `liveflux_component_id`.
2. `Handler.mount()` creates a new component instance (`newByAlias(alias)`), generates an ID (`NewID()`), and calls `Mount` with filtered params.
3. The component is persisted via `Store.Set` and rendered. The HTML is returned with `Content-Type: text/html; charset=utf-8`.

### Action Requests

1. Clients submit forms including `liveflux_component_id` and optionally `liveflux_action`.
2. The handler loads the component from the store, calls `Handle` when an action is provided, persists the updated state, and re-renders HTML.
3. Redirects requested during `Handle` are honored (see below).

### Error Handling

- Missing alias or ID → `400 Bad Request`.
- Unknown alias or missing component → `404 Not Found`.
- `Mount`/`Handle` returning an error → `500`/`400`, plus a log line (`log.Printf`).

## Redirects

Components can call `Base.Redirect(url, delaySeconds...)`. The handler reads redirect metadata through `TakeRedirect()` and `TakeRedirectDelaySeconds()`, then:

- Sets `X-Liveflux-Redirect` and optionally `X-Liveflux-Redirect-After` headers.
- Writes a minimal HTML body containing a JavaScript redirect with `<noscript>` fallback.

Clients shipped with Liveflux automatically consume these headers.

## Transport Options

### HTTP Only

Use when WebSockets are unnecessary:

```go
mux.Handle("/liveflux", liveflux.NewHandler(nil))
```

### Combined HTTP + WebSocket

`liveflux.NewHandlerWS(store)` delegates to `NewWebSocketHandler`. It automatically upgrades WebSocket requests and falls back to HTTP for others. The default client can mount over HTTP and optionally negotiate WebSockets later.

### Explicit Choice

`liveflux.NewHandlerEx(store, enableWebSocket)` toggles WebSocket support at runtime.

## Middleware Integration

Wrap the handler to add middleware:

```go
mux.Handle("/liveflux", myAuthMiddleware(liveflux.NewHandler(store)))
```

Ensure middleware does not consume the request body or mutate form fields before the handler runs.

## Custom Routers

Because Liveflux exposes standard `http.Handler`, it works with routers like chi, gin, echo, or fiber. Example with chi:

```go
r := chi.NewRouter()
r.Post("/liveflux", liveflux.NewHandler(nil).ServeHTTP)
```

For frameworks requiring handler functions rather than `http.Handler`, wrap the handler’s `ServeHTTP` method.

## CSRF and Security

- Add CSRF tokens via `ClientOptions.Headers` or hidden form inputs.
- Validate tokens inside `Handle`. For WebSockets, use `WithWebSocketCSRFCheck` to inspect the upgrade request before accepting it.
- Restrict allowed methods by hosting the handler under a path protected by router-level middleware.

## Custom Stores

Handlers interact with the configured `Store` exclusively via the interface methods. Implement custom stores when you need persistence across processes, e.g., to share state between replicas.

## Testing Handlers

Simulate HTTP requests with `httptest`:

```go
req := httptest.NewRequest("POST", "/liveflux", strings.NewReader(form.Encode()))
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
rr := httptest.NewRecorder()
handler := liveflux.NewHandler(liveflux.NewMemoryStore())
handler.ServeHTTP(rr, req)
```

Use `handler_test.go` as a reference for expected status codes and behaviors.
