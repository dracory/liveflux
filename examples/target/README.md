# Targeted Fragment Updates Example

This example demonstrates Liveflux's **targeted fragment updates** feature (`data-flux-target`), which allows components to update only specific DOM regions instead of re-rendering the entire component.

## Features Demonstrated

- **Selective Updates**: Only the cart total and item list are updated when you add/remove items
- **Dirty Tracking**: Components mark specific targets as dirty using `MarkTargetDirty()`
- **TargetRenderer Interface**: Components implement `RenderTargets()` to return only changed fragments
- **Automatic Fallback**: Full component render is included as fallback if targets fail

## How It Works

### 1. Component Implements TargetRenderer

```go
func (c *CartComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
    fragments := []liveflux.TargetFragment{}
    
    if c.IsDirty("#cart-total") {
        fragments = append(fragments, liveflux.TargetFragment{
            Selector: "#cart-total",
            Content:  c.renderTotal(),
            SwapMode: liveflux.SwapReplace,
        })
    }
    
    return fragments
}
```

### 2. Actions Mark Targets Dirty

```go
func (c *CartComponent) Handle(ctx context.Context, action string, data url.Values) error {
    switch action {
    case "add-item":
        // ... modify state ...
        c.MarkTargetDirty("#cart-total")
        c.MarkTargetDirty(".line-items")
    }
    return nil
}
```

### 3. Client Receives Template Response

Instead of full HTML, the server sends:

```html
<template data-flux-target="#cart-total" data-flux-swap="replace" data-flux-component-kind="cart" data-flux-component-id="abc123">
  <div id="cart-total">Total: $125.00</div>
</template>

<template data-flux-target=".line-items" data-flux-swap="replace" data-flux-component-kind="cart" data-flux-component-id="abc123">
  <ul class="line-items">...</ul>
</template>

<template data-flux-component-kind="cart" data-flux-component-id="abc123">
  <!-- Full component render as fallback -->
</template>
```

### 4. Client Applies Fragments

The client JavaScript:
1. Detects template-based response
2. Applies each fragment to its selector
3. Falls back to full render if any selector fails

## Running the Example

```bash
cd examples/target
go run main.go
```

Open http://localhost:8080 and:
1. Open browser DevTools Network tab
2. Click "Add Item" or "Remove Item"
3. Observe that the response contains only `<template>` elements, not full HTML
4. Notice only the total and item list update in the DOM

## Benefits

- **Reduced Network Payload**: Only changed fragments are sent
- **Preserved Client State**: Focus, scroll position, and third-party widgets remain intact
- **Better Performance**: Less HTML parsing and DOM manipulation
- **Smooth Animations**: CSS transitions work correctly since elements are updated, not replaced

## Swap Modes

The example uses `SwapReplace`, but other modes are available:

- `liveflux.SwapReplace` - Replace the target element entirely
- `liveflux.SwapInner` - Replace the target's innerHTML
- `liveflux.SwapBeforeBegin` - Insert before the target
- `liveflux.SwapAfterBegin` - Insert as first child
- `liveflux.SwapBeforeEnd` - Insert as last child (append)
- `liveflux.SwapAfterEnd` - Insert after the target

## See Also

- [data-flux-target Proposal](../../docs/proposals/data-flux-target.md) - Full specification
- [Counter Example](../counter/) - Basic component without targets
- [CRUD Example](../crud/) - More complex component patterns
