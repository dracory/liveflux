# Getting Started with Liveflux

This guide walks through the process of wiring Liveflux into a Go web application. It mirrors the quick-start in `README.md` with detailed explanations and variations.

## Prerequisites

- Go 1.22 or newer (matches `go.mod` requirements)
- Basic familiarity with `net/http`
- Access to the Liveflux module (`go get github.com/dracory/liveflux`)

## 1. Register a Component

Create a component struct that satisfies `liveflux.ComponentInterface`. Embed `liveflux.Base` to inherit alias/ID handling and helper methods.

```go
package counter

import (
    "context"
    "net/url"

    "github.com/dracory/hb"
    "github.com/dracory/liveflux"
)

type Component struct {
    liveflux.Base
    Count int
}

func (c *Component) GetAlias() string { return "counter" }

func (c *Component) Mount(ctx context.Context, params map[string]string) error {
    c.Count = 0
    return nil
}

func (c *Component) Handle(ctx context.Context, action string, form url.Values) error {
    switch action {
    case "inc":
        c.Count++
    case "dec":
        c.Count--
    }
    return nil
}

func (c *Component) Render(ctx context.Context) hb.TagInterface {
    return c.Root(hb.Div().Textf("%d", c.Count))
}

func init() {
    _ = liveflux.Register(new(Component))
}
```

- `Mount` initializes state with optional params from placeholders.
- `Handle` reacts to actions sent by the client.
- `Render` builds HTML using `hb` (or any `hb.TagInterface`).
- `c.Root(...)` wraps the content with hidden inputs for alias and ID.

## 2. Expose the Endpoint

Add a route targeting `liveflux.NewHandler(store)`. Passing `nil` uses the default in-memory store (`StoreDefault`).

```go
mux := http.NewServeMux()
mux.Handle("/liveflux", liveflux.NewHandler(nil))
```

Use `liveflux.NewHandlerWS(nil)` to automatically upgrade WebSocket requests while falling back to HTTP when needed.

## 3. Render Server HTML (Optional SSR)

Server-side rendering produces the initial markup:

```go
counter1, _ := liveflux.New(&counter.Component{})
counter2, _ := liveflux.New(&counter.Component{})

layout := hb.HTML(
    hb.Head(
        hb.Title().Text("Counters"),
    ),
    hb.Body(
        liveflux.SSR(counter1),
        liveflux.SSR(counter2),
        liveflux.Script(),
    ),
)

w.Header().Set("Content-Type", "text/html; charset=utf-8")
_, _ = w.Write([]byte(layout.ToHTML()))
```

`liveflux.SSR()` mounts the component, stores it in `StoreDefault`, and returns hydrated markup that includes the required hidden inputs.

## 4. Include the Client Runtime

The client script must be present once per page. If you are not using `hb`, emit the string returned by `liveflux.Script()` manually:

```go
fmt.Fprintf(w, "<script>%s</script>", liveflux.JS())
```

Configure the client via `liveflux.ClientOptions`:

```go
liveflux.Script(liveflux.ClientOptions{
    Endpoint:    "/api/liveflux",
    Credentials: "include",
    Headers:     map[string]string{"X-CSRF-Token": token},
    UseWebSocket: true,
})
```

## 5. Mount from HTML

Add placeholders where components should appear. The client picks them up by `data-flux-mount="1"` and posts to the endpoint.

```go
placeholder := liveflux.PlaceholderByAlias("counter", map[string]string{"theme": "dark"})
```

Rendered HTML:

```html
<div data-flux-mount="1" data-flux-component="counter" data-flux-param-theme="dark">Loading counter...</div>
```

The `theme` parameter becomes `params["theme"]` in `Mount`.

## Running the Example

Clone the repository and run the counter demo:

```bash
go run ./examples/counter
```

Visit `http://localhost:8080` to interact with two counters sharing the same endpoint and store.

## Next Steps

- Review `docs/components.md` for best practices in component design.
- Consult `docs/handler_and_transport.md` for transport configuration.
- Explore `docs/ssr.md` and `docs/websocket.md` for advanced rendering and realtime integrations.
