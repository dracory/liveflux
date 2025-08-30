# Livewire-style Components for Go (with hb)

Server-driven, Livewire-like components for Go. Renders HTML with [`github.com/gouniverse/hb`](https://github.com/gouniverse/hb) and updates via small HTMX requests that swap only the component DOM.

- Endpoint: `POST /livewire`
- Transport: [HTMX](https://htmx.org/) (recommended)
- Rendering: `hb.Tag.ToHTML()`
- State: In-memory by default (swap with session-backed store for production)

## Quick start

1) Register a component

```go
// internal/components/counter/counter.go
package counter

import (
    "context"
    "net/url"

    "github.com/gouniverse/hb"
    "project/pkg/livewire"
)

type Component struct { livewire.Base; Count int }

func (c *Component) Mount(ctx context.Context, params map[string]string) error {
    c.Count = 0
    return nil
}

func (c *Component) Handle(ctx context.Context, action string, data url.Values) error {
    switch action {
    case "inc": c.Count++
    case "dec": c.Count--
    case "reset": c.Count = 0
    }
    return nil
}

func (c *Component) Render(ctx context.Context) hb.Tag {
    root := hb.Div().
        Attr("hx-post", "/livewire").
        Attr("hx-target", "this").
        Attr("hx-swap", "outerHTML").
        Child(hb.Input().Type("hidden").Name(livewire.FormAlias).Value("counter")).
        Child(hb.Input().Type("hidden").Name(livewire.FormID).Value(c.ID()))

    root.Child(hb.H2().Text("Counter"))
    root.Child(hb.Div().Style("font-size:2rem").Textf("%d", c.Count))
    root.Child(hb.Button().Text("+1").Attr("hx-vals", `{"action":"inc"}`))
    root.Child(hb.Button().Text("-1").Attr("hx-vals", `{"action":"dec"}`))
    root.Child(hb.Button().Text("Reset").Attr("hx-vals", `{"action":"reset"}`))
    return *root
}

func init() { livewire.Register("counter", func() livewire.ComponentInterface { return &Component{} }) }
```

2) Wire the endpoint

`/livewire` is already registered by `internal/controllers/livewire/routes.go` and included in `internal/routes/routes.go`.

3) Mount the component (HTMX)

Raw HTML:

```html
<div hx-post="/livewire" hx-trigger="load" hx-target="this" hx-swap="outerHTML">
  <input type="hidden" name="alias" value="counter" />
</div>
```

With `hb`:

```go
hb.Div().
  Attr("hx-post", "/livewire").
  Attr("hx-trigger", "load").
  Attr("hx-target", "this").
  Attr("hx-swap", "outerHTML").
  Child(hb.Input().Type("hidden").Name("alias").Value("counter"))
```

After the first response, the component HTML contains hidden `alias` and `id`, so subsequent button clicks only need to send `action` via `hx-vals`.

## Package API

- Interface `livewire.ComponentInterface`:
  - `ID() string`, `SetID(id string)`
  - `Mount(ctx, params map[string]string) error`
  - `Handle(ctx, action string, data url.Values) error`
  - `Render(ctx) hb.Tag`
- Registry:
  - `livewire.Register(alias string, ctor func() ComponentInterface)`
  - `livewire.New(component ComponentInterface) (ComponentInterface, error)`
- Handler:
  - `livewire.NewHandler(store Store)` → `http.Handler`
- Form fields (constants):
  - `alias`, `id`, `action`

## State store

Default: in-memory `MemoryStore` (process-local). For multi-instance or restart-safe state, provide your own implementation of `livewire.Store` (e.g., session-backed) and pass it to `livewire.NewHandler(store)`.

## Production notes

- Validate inputs in `Handle`.
- Avoid storing secrets in component fields if using client-side state approaches.
- For larger UIs, split into nested components and mount them independently.
- Add CSRF as needed (HTMX posts are standard forms).

## License
MIT
