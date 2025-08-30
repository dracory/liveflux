# Liveflux — Server-driven UI Components for Go

[![Tests Status](https://github.com/dracory/liveflux/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/dracory/liveflux/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/liveflux)](https://goreportcard.com/report/github.com/dracory/liveflux)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/liveflux)](https://pkg.go.dev/github.com/dracory/liveflux)

Liveflux is a server-driven component system for Go. It uses [`github.com/gouniverse/hb`](https://github.com/gouniverse/hb) internally for component tags, but it works with any server setup or frontend—responses are plain HTML and the client transport is framework-agnostic.

- Endpoint: your choice (example: `POST /liveflux`) (handler accepts `POST` and `GET`)
- Transport: built-in JS (any client that can POST/GET forms works)
- Rendering: `hb.TagInterface.ToHTML()` (internal detail). You can integrate with any templating/system since the handler returns HTML.
- State: In-memory by default via `MemoryStore` (swap with a session-backed store for production)

## Quick start

1) Register a component

```go
// internal/components/counter/counter.go
package counter

import (
    "context"
    "fmt"
    "net/url"

    "github.com/gouniverse/hb"
    "github.com/dracory/liveflux"
)

type Component struct { liveflux.Base; Count int }

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

func (c *Component) Render(ctx context.Context) hb.TagInterface {
    root := hb.Div().
        Attr("hx-post", "/liveflux"). // example endpoint
        Attr("hx-target", "this").
        Attr("hx-swap", "outerHTML").
        // Handler expects form fields: component, id, action
        Child(hb.Input().Type("hidden").Name("component").Value("counter")).
        Child(hb.Input().Type("hidden").Name("id").Value(c.GetID()))

    root = root.Child(hb.H2().Text("Counter"))
    root = root.Child(hb.Div().Style("font-size:2rem").Text(fmt.Sprintf("%d", c.Count)))
    root = root.Child(hb.Button().Text("+1").Attr("hx-vals", `{"action":"inc"}`))
    root = root.Child(hb.Button().Text("-1").Attr("hx-vals", `{"action":"dec"}`))
    root = root.Child(hb.Button().Text("Reset").Attr("hx-vals", `{"action":"reset"}`))
    return root
}

// Register using an alias
func init() { liveflux.RegisterByAlias("counter", func() liveflux.ComponentInterface { return &Component{} }) }
```

2) Wire the endpoint

Create an HTTP route and attach the handler:

```go
// main.go (or your router setup)
package main

import (
    "net/http"
    "github.com/dracory/liveflux"
)

func main() {
    mux := http.NewServeMux()
    mux.Handle("/liveflux", liveflux.NewHandler(nil)) // nil -> uses default in-memory store
    http.ListenAndServe(":8080", mux)
}
```

3) Mount the component (HTMX)

Raw HTML:

```html
<div hx-post="/liveflux" hx-trigger="load" hx-target="this" hx-swap="outerHTML">
  <input type="hidden" name="component" value="counter" />
</div>
```

With `hb`:

```go
hb.Div().
  Attr("hx-post", "/liveflux").
  Attr("hx-trigger", "load").
  Attr("hx-target", "this").
  Attr("hx-swap", "outerHTML").
  Child(hb.Input().Type("hidden").Name("component").Value("counter"))
```

After the first response, the component HTML contains hidden `component` and `id`, so subsequent button clicks only need to send `action` via `hx-vals`.

## Package API

- Interface `liveflux.ComponentInterface`:
  - `GetID() string`, `SetID(id string)`
  - `Mount(ctx, params map[string]string) error`
  - `Handle(ctx, action string, data url.Values) error`
  - `Render(ctx) hb.TagInterface`
- Registry:
  - `liveflux.RegisterByAlias(alias string, ctor func() ComponentInterface)`
  - `liveflux.Register(ctor func() ComponentInterface)`
  - `liveflux.New(example ComponentInterface) (ComponentInterface, error)`
- Handler:
  - `liveflux.NewHandler(store Store)` → `http.Handler`
- Form fields (constants):
  - `component`, `id`, `action`

Constants exported by the package:

- `FormComponent = "component"`
- `FormID = "id"`
- `FormAction = "action"`

### SSR (server-side render) helpers

- `liveflux.SSR(c ComponentInterface, params ...map[string]string) hb.TagInterface`
- `liveflux.SSRHTML(c ComponentInterface, params ...map[string]string) string`

Use these to mount and render a component entirely on the server once (good for SEO) while still enabling the client runtime to hydrate later.

Example:

```go
// server-side
html := liveflux.SSRHTML(&counter.Component{}, map[string]string{"userID": "42"})
// write `html` into your page; include client JS to hydrate for actions
```

### Placeholders and client script

- `liveflux.PlaceholderByAlias(alias string, params ...map[string]string) hb.TagInterface`
- `liveflux.Placeholder(c ComponentInterface, params ...map[string]string) hb.TagInterface`
- `liveflux.JS() string` and `liveflux.Script() hb.TagInterface` return the minimal client JS required for mounting and actions.

Include the client once per page (layout):

```go
// Using hb
layout = layout.Child(liveflux.Script())
// or: layout = layout.Child(hb.Script(liveflux.JS()))
```

Mount via placeholder (built-in client auto-mounts on load):

```go
// Placeholder by alias; optional params become data attributes
ph := liveflux.PlaceholderByAlias("counter", map[string]string{"theme": "dark"})
// Renders: <div data-lw-mount="1" data-lw-component="counter" data-lw-param-theme="dark">Loading counter...</div>
```

Data attributes used by the client:

- `data-lw-mount="1"` — element to mount
- `data-lw-component="<alias>"` — registered alias
- `data-lw-param-<name>="<value>"` — initial params passed to `Mount`

HTMX remains fully compatible; use whichever transport you prefer.

## State store

Default: in-memory `MemoryStore` (process-local), exposed as `StoreDefault`. For multi-instance or restart-safe state, provide your own implementation of `liveflux.Store` (e.g., session-backed) and pass it to `liveflux.NewHandler(store)`.

Registry & aliases:

- `RegisterByAlias(alias string, ctor func() ComponentInterface)`
- `Register(ctor func() ComponentInterface)` — uses `GetAlias()` or derives via `DefaultAliasFromType`
- `New(example ComponentInterface)` — constructs a fresh instance using the example's alias/type

`DefaultAliasFromType` derives `<pkg>.<type-kebab>` or just `<pkg>` when names match.

Redirects:

- If a component calls `Base.Redirect(url, delaySeconds...)`, the handler sets custom redirect headers and writes a small HTML fallback that performs the redirect via `<script>` and `<noscript>` meta refresh.

## Production notes

- Validate inputs in `Handle`.
- Avoid storing secrets in component fields if using client-side state approaches.
- For larger UIs, split into nested components and mount them independently.
- Add CSRF as needed (HTMX posts are standard forms).

## License

This project is dual-licensed under the following terms:

- For non-commercial use, you may choose either the GNU Affero General Public License v3.0 (AGPLv3) _or_ a separate commercial license (see below). You can find a copy of the AGPLv3 at: https://www.gnu.org/licenses/agpl-3.0.txt

- For commercial use, a separate commercial license is required. Commercial licenses are available for various use cases. Please contact me via my [contact page](https://lesichkov.co.uk/contact) to obtain a commercial license.

## Similar projects

- Laravel Livewire (PHP): https://livewire.laravel.com
- Phoenix LiveView (Elixir): https://hexdocs.pm/phoenix_live_view
- Hotwire Turbo (Rails): https://turbo.hotwired.dev
- StimulusReflex (Ruby): https://docs.stimulusreflex.com
- Inertia.js (multi-framework): https://inertiajs.com
- Blazor (C#/.NET): https://dotnet.microsoft.com/apps/aspnet/web-apps/blazor
