# Components

Components are the core building blocks in Liveflux. Each component is a Go struct implementing `liveflux.ComponentInterface` and responsible for its own state, lifecycle, and HTML rendering.

## Lifecycle Overview

1. **Construction**: The framework instantiates a component by kind through the registry (`registry.go`). `liveflux.Base.SetKind()` ensures the kind is set once.
2. **Mount**: `Mount(ctx, params)` initializes state. `params` are collected from placeholder data attributes or SSR inputs.
3. **Render**: `Render(ctx)` returns an `hb.TagInterface` describing the current UI.
4. **Handle**: When the client triggers an action, `Handle(ctx, action, form)` mutates state.
5. **Re-render**: After `Handle`, the framework calls `Render` again and returns the HTML diff to the client.

## Implementing a Component

```go
type TodoList struct {
    liveflux.Base
    Items []string
}

func (c *TodoList) GetKind() string { return "todo.list" }

func (c *TodoList) Mount(ctx context.Context, params map[string]string) error {
    if initial := params["initial"]; initial != "" {
        c.Items = strings.Split(initial, ",")
    }
    return nil
}

func (c *TodoList) Handle(ctx context.Context, action string, vals url.Values) error {
    switch action {
    case "add":
        c.Items = append(c.Items, vals.Get("item"))
    case "remove":
        idx, _ := strconv.Atoi(vals.Get("idx"))
        if idx >= 0 && idx < len(c.Items) {
            c.Items = append(c.Items[:idx], c.Items[idx+1:]...)
        }
    }
    return nil
}

func (c *TodoList) Render(ctx context.Context) hb.TagInterface {
    list := hb.Ul()
    for i, item := range c.Items {
        list = list.Child(
            hb.Li().Text(item).
                Child(hb.Button().Data("flux-action", "remove").Data("flux-field-idx", strconv.Itoa(i)).Text("✕")),
        )
    }

    form := hb.Form().
        Method("post").
        Child(hb.Input().Name("item").Placeholder("Add item"))).
        Child(hb.Button().Type("submit").Data("flux-action", "add").Text("Add"))

    return c.Root(hb.Div().Child(list).Child(form))
}

func init() { _ = liveflux.Register(new(TodoList)) }
```

### Tips

- Always call `c.Root(...)` to include `data-flux-kind` and `data-flux-component-id` attributes.
- Keep `Render` deterministic based on component fields. Avoid non-idempotent side effects.
- Validate user input in `Handle` to maintain server trust.

## Parameter Handling

Placeholder attributes like `data-flux-param-theme="dark"` map to `params["theme"]` in `Mount`. Use this to pass initial state or configuration.

## Form-less Field Collection

Components can collect form data from anywhere in the DOM using `data-flux-include` and `data-flux-exclude` attributes on action buttons:

```go
hb.Button().
    Attr(liveflux.DataFluxAction, "save").
    Attr(liveflux.DataFluxInclude, "#external-form").
    Text("Save")
```

This allows sharing form inputs across multiple components without requiring traditional `<form>` wrappers. The client runtime serializes all fields from the specified selectors and includes them in the action request. See `docs/handler_and_transport.md` for detailed usage and `examples/formless/` for working examples.

## Response Fragment Filtering (`data-flux-select`)

Use `data-flux-select` on triggers (buttons, links, forms) when the server returns a full HTML document but only a fragment should replace the component root:

```html
<button
  data-flux-action="refresh"
  data-flux-select="#cart-summary, #cart-summary > .total:first-of-type">
  Refresh Cart
</button>
```

- The client parses the response with `DOMParser`, evaluates selectors in order, and swaps with the first match.
- When no selector matches, Liveflux logs a warning and falls back to the original HTML to avoid blank updates.
- The attribute works with both click actions and form submissions. For forms without a submitter (e.g., programmatic submission), set `data-flux-select` on the `<form>` element itself.
- To inspect selection decisions during development, set `window.liveflux.debugSelect = true` in the browser console to print verbose logs.

## Redirects

`liveflux.Base` exposes redirect helpers consumed by `Handler`:

```go
c.Redirect("/dashboard")
c.Redirect("/thanks", 5) // delay in seconds
```

`Handler` (`handler.go`) converts these into headers (`RedirectHeader`, `RedirectAfterHeader`) and fallback HTML.

## Nested Components

Compose UIs by embedding placeholders in your component’s `Render` output:

```go
hb.Div().
    Child(liveflux.PlaceholderByKind("notifications")).
    Child(liveflux.PlaceholderByKind("chat.panel"))
```

Each nested component mounts independently and maintains its own state. Use distinct kinds to prevent collisions.

## Testing Components

Write unit tests against `Mount`, `Handle`, and `Render` without the handler:

```go
func TestTodoList_Handle_Add(t *testing.T) {
    c := &TodoList{}
    _ = c.Mount(context.Background(), nil)
    _ = c.Handle(context.Background(), "add", url.Values{"item": {"Test"}})
    if len(c.Items) != 1 {
        t.Fatalf("expected 1 item")
    }
}
```

Use `component_test.go` as a reference for verifying `liveflux.Base` behavior.
