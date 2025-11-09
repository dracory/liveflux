# Handler and Transport

Liveflux exposes HTTP handlers that integrate with any Go router. This guide explains how the request pipeline works, how to configure transports, and how to integrate with existing middleware.

## HTTP Handler (`handler.go`)

`liveflux.NewHandler(store Store) *Handler` returns an `http.Handler` that accepts `POST` and `GET` requests. `GET` responses stream the embedded client runtime so you can serve `<script src="/liveflux" defer></script>` directly from the same endpoint. `POST` requests expect the following form fields:

- `liveflux_component_alias` (`FormComponent` constant): component alias.
- `liveflux_component_id` (`FormID`): assigned during mount; required for actions.
- `liveflux_action` (`FormAction`): optional action identifier.

All other form fields are passed to the component's `Handle` method as `url.Values`.

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

Include the client on your pages by linking to the same path:

```html
<script src="/liveflux" defer></script>
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

## Form-less Submission

The client runtime supports flexible field collection using `data-flux-include` and `data-flux-exclude` attributes on action buttons. This allows components to collect data from arbitrary DOM elements without requiring traditional `<form>` wrappers.

### Basic Usage

Add `data-flux-include` to an action button to collect fields from elements outside the component root:

```html
<button data-flux-action="save" data-flux-include="#user-form">
  Save
</button>
```

The client will serialize all input fields within `#user-form` and include them in the action request.

### Multiple Selectors

Specify multiple CSS selectors separated by commas:

```html
<button data-flux-action="submit" 
        data-flux-include="#step-1, #step-2, .shared-inputs">
  Submit
</button>
```

### Excluding Fields

Use `data-flux-exclude` to omit specific fields from included scopes:

```html
<button data-flux-action="update" 
        data-flux-include="#profile-form"
        data-flux-exclude=".sensitive">
  Update Profile
</button>
```

This includes all fields from `#profile-form` except those with the `sensitive` class.

### Field Precedence

When the same field name appears in multiple sources, the precedence (lowest to highest) is:

1. Component root fields
2. Associated form fields
3. Included elements (left to right in selector list)
4. Excluded elements (removed)
5. Button `data-flux-param-*` attributes
6. Button `name`/`value` (if applicable)

### Server-Side Helpers

Use the `IncludeSelectors` and `ExcludeSelectors` helpers to build attribute values:

```go
hb.Button().
    Attr(liveflux.DataFluxAction, "save").
    Attr(liveflux.DataFluxInclude, liveflux.IncludeSelectors("#form-1", "#form-2")).
    Text("Save")
```

### Benefits

- **Progressive enhancement**: Works without JavaScript by using non-form containers
- **Flexible composition**: Share form fragments across multiple components
- **Reduced overhead**: No need for `<form>` wrappers everywhere
- **Fine-grained control**: Include or exclude specific fields as needed

See `examples/formless/` for complete working examples.

### Request Indicators

Liveflux mirrors htmx-style loading indicators. Add `data-flux-indicator` (or `flux-indicator`) to any trigger element (button, link, mount placeholder) and set it to a CSS selector list. Every request started from that trigger toggles the `flux-request` and `htmx-request` classes on the referenced elements for the duration of the network call. When the attribute is omitted, Liveflux falls back to any `.flux-indicator` or `.htmx-indicator` elements inside the component root or on the trigger itself.

Example:

```html
<button data-flux-action="save" data-flux-indicator="#spinner">
  Save
  <span id="spinner" class="hidden">Saving…</span>
</button>
```

```css
.hidden { display: none; }
.hidden.flux-request { display: inline-block; }
```

You can also use the literal value `this` to toggle the trigger element. The compatibility `htmx-request` class makes existing htmx indicator styles work without changes.
