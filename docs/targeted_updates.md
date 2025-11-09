# Targeted Fragment Updates

Liveflux supports **targeted fragment updates**, allowing components to update only specific DOM regions instead of re-rendering the entire component. This reduces network payloads, preserves client-side state (focus, scroll, video playback), and enables more granular UI composition.

## Quick Start

### 1. Implement TargetRenderer

```go
type CartComponent struct {
    liveflux.Base
    Total float64
    Items []CartItem
}

func (c *CartComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
    fragments := []liveflux.TargetFragment{}
    
    // Only render fragments for dirty targets
    if c.IsDirty("#cart-total") {
        fragments = append(fragments, liveflux.TargetFragment{
            Selector: "#cart-total",
            Content:  c.renderTotal(),
            SwapMode: liveflux.SwapReplace,
        })
    }

    if c.IsDirty("#global-cart-badge") {
        fragments = append(fragments, liveflux.TargetFragment{
            Selector:            "#global-cart-badge",
            Content:             c.renderGlobalBadge(),
            SwapMode:            liveflux.SwapInner,
            NoComponentMetadata: true, // document-scoped fragment
        })
    }
    
    return fragments
}
```

### 2. Mark Targets Dirty in Actions

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

### 3. Use Stable Selectors in Render

```go
func (c *CartComponent) Render(ctx context.Context) hb.TagInterface {
    return c.Root(
        hb.Div().Children([]hb.TagInterface{
            hb.Div().ID("cart-total").Text(fmt.Sprintf("$%.2f", c.Total)),
            hb.UL().Class("line-items").Children(c.renderItems()),
        }),
    )
}
```

## How It Works

### Document-scoped fragments (global selectors)

Fragments without `data-flux-component` metadata are treated as document scoped. Set `NoComponentMetadata` to `true` when you want to patch DOM outside the component root, such as a header badge shared across multiple components:

```go
if c.IsDirty("#global-cart-badge") {
    fragments = append(fragments, liveflux.TargetFragment{
        Selector:            "#global-cart-badge",
        Content:             c.renderGlobalBadge(),
        SwapMode:            liveflux.SwapInner,
        NoComponentMetadata: true,
    })
}
```

On the client, the fragment selector is resolved against `document` instead of the component root, so the badge updates even though it lives outside the cart component tree. Keep selectors unique to avoid accidental matches.

### Client-Server Flow

1. **Client sends request** with `X-Liveflux-Target: enabled` header (auto-enabled by default)
2. **Server checks** if component implements `TargetRenderer`
3. **Server returns** template-based response with fragments:
   ```html
   <template data-flux-target="#cart-total" data-flux-swap="replace">
     <div id="cart-total">$125.00</div>
   </template>
   <template data-flux-component="cart" data-flux-component-id="abc123">
     <!-- Full component render as fallback -->
   </template>
   ```
4. **Client applies** each fragment to its selector
5. **Client falls back** to full render if any selector fails

### Automatic Handshake

The client automatically enables target support by adding the `X-Liveflux-Target: enabled` header to all requests. The server detects this header and returns template-based responses when the component implements `TargetRenderer`.

To disable target support:
```javascript
liveflux.disableTargetSupport();
```

To re-enable:
```javascript
liveflux.enableTargetSupport();
```

## Swap Modes

Control how fragments are merged with target elements:

| Mode | Constant | Description |
|------|----------|-------------|
| `replace` | `liveflux.SwapReplace` | Replace the target element entirely (default) |
| `inner` | `liveflux.SwapInner` | Replace the target's innerHTML |
| `beforebegin` | `liveflux.SwapBeforeBegin` | Insert before the target element |
| `afterbegin` | `liveflux.SwapAfterBegin` | Insert as first child of target |
| `beforeend` | `liveflux.SwapBeforeEnd` | Insert as last child of target (append) |
| `afterend` | `liveflux.SwapAfterEnd` | Insert after the target element |

### Example: Appending Items

```go
func (c *ListComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
    if c.IsDirty(".items") {
        return []liveflux.TargetFragment{{
            Selector: ".items",
            Content:  c.renderNewItem(),
            SwapMode: liveflux.SwapBeforeEnd, // Append to list
        }}
    }
    return nil
}
```

## Dirty Tracking

The `Base` struct provides dirty tracking helpers:

```go
// Mark a target as needing an update
c.MarkTargetDirty("#cart-total")

// Check if a target is dirty
if c.IsDirty("#cart-total") {
    // render fragment
}

// Get all dirty targets
dirty := c.GetDirtyTargets() // map[string]bool

// Clear all dirty markers
c.ClearDirtyTargets()
```

## Selector Helpers

Build selectors programmatically:

```go
// ID selector
liveflux.TargetID("my-element")        // "#my-element"

// Class selector
liveflux.TargetClass("my-class")       // ".my-class"

// Attribute selector
liveflux.TargetAttr("data-id", "42")   // "[data-id='42']"

// Custom selector
liveflux.TargetSelector("#custom")     // "#custom"
```

## Component Metadata Validation

Fragments can include component metadata for validation:

```go
liveflux.TargetFragment{
    Selector: "#total",
    Content:  hb.Div().Text("$125"),
    SwapMode: liveflux.SwapReplace,
}
```

The server automatically adds `data-flux-component` and `data-flux-component-id` attributes to each template. The client validates these before applying fragments, preventing cross-component updates.

## WebSocket Support

Targeted updates work seamlessly with WebSocket transport. The WebSocket handler detects template responses and applies fragments just like HTTP requests:

```go
// No special code needed - works automatically
handler := liveflux.NewHandlerWS(nil)
```

## Fallback Behavior

### Server-Side

If a component implements `TargetRenderer` but returns no fragments, the server falls back to full component rendering:

```go
func (c *Component) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
    if !c.hasChanges() {
        return nil // Server will send full render
    }
    return c.buildFragments()
}
```

### Client-Side

If any selector fails to match, the client falls back to full component replacement:

```javascript
// Automatic fallback - no code needed
// Client logs warning and uses full render from fallback template
```

## Best Practices

### 1. Use Stable Selectors

Ensure selectors remain consistent across renders:

```go
// Good: Stable ID
hb.Div().ID("cart-total").Text(...)

// Bad: Dynamic ID that changes
hb.Div().ID(fmt.Sprintf("total-%d", time.Now().Unix())).Text(...)
```

### 2. Mark Targets Immediately

Mark targets dirty as soon as state changes:

```go
func (c *Component) Handle(ctx context.Context, action string, data url.Values) error {
    c.Total += 10.00
    c.MarkTargetDirty("#total") // Mark immediately
    return nil
}
```

### 3. Render Fragments Independently

Each fragment should be self-contained:

```go
func (c *Component) renderTotal() hb.TagInterface {
    // Complete element with all attributes
    return hb.Div().
        ID("cart-total").
        Class("total").
        Text(fmt.Sprintf("$%.2f", c.Total))
}
```

### 4. Include Full Render Fallback

The server automatically includes a full render fallback in every response. Ensure your `Render()` method always returns complete, valid HTML.

### 5. Clear Dirty Targets After Render

Dirty targets are automatically cleared after `RenderTargets()` is called. If you need manual control, use `ClearDirtyTargets()`.

## Performance Considerations

### Network Savings

Targeted updates can reduce response sizes by 90%+ for large components:

```
Full render:     15 KB
Targeted update: 1.2 KB (92% reduction)
```

### When to Use Targets

**Use targeted updates when:**
- Component has large, mostly static sections
- Only small portions change frequently
- Preserving client state is important (focus, scroll, media playback)

**Use full renders when:**
- Entire component changes on every action
- Component is small (<2 KB)
- Simplicity is more important than optimization

## Debugging

### Enable Logging

The client logs all target operations to the console:

```
[Liveflux Target] Applied: #cart-total (mode: replace)
[Liveflux Target] Applied: .line-items (mode: replace)
```

### Check Network Tab

Inspect responses in browser DevTools Network tab:
1. Look for `<template>` elements in response
2. Verify `X-Liveflux-Target: enabled` header in request
3. Check response size compared to full render

### Common Issues

**Selector not found:**
- Verify selector matches element in DOM
- Check for typos in ID/class names
- Ensure element exists before fragment is applied

**Component mismatch:**
- Verify component kind matches
- Check component ID is correct
- Ensure fragment targets correct component instance

**Fallback triggered:**
- Check console for warnings
- Verify all selectors are valid
- Ensure fragments are not empty

## Examples

See the [target example](../examples/target/) for a complete working implementation.

## API Reference

### Go

```go
// Interface
type TargetRenderer interface {
    RenderTargets(ctx context.Context) []TargetFragment
}

// Types
type TargetFragment struct {
    Selector string
    Content  hb.TagInterface
    SwapMode string
}

// Base methods
func (b *Base) MarkTargetDirty(selector string)
func (b *Base) IsDirty(selector string) bool
func (b *Base) ClearDirtyTargets()
func (b *Base) GetDirtyTargets() map[string]bool

// Helpers
func TargetID(id string) string
func TargetClass(class string) string
func TargetAttr(attr, value string) string
func TargetSelector(selector string) string
func BuildTargetResponse(fragments []TargetFragment, fullRender string, comp ComponentInterface) string

// Constants
const SwapReplace = "replace"
const SwapInner = "inner"
const SwapBeforeBegin = "beforebegin"
const SwapAfterBegin = "afterbegin"
const SwapBeforeEnd = "beforeend"
const SwapAfterEnd = "afterend"
```

### JavaScript

```javascript
// Functions
liveflux.applyTargets(html, componentRoot)
liveflux.hasTargetTemplates(html)
liveflux.enableTargetSupport()
liveflux.disableTargetSupport()

// Configuration
liveflux.autoEnableTargets = true // Default
```

## Migration Guide

### From Full Renders

1. **Identify frequently updated sections** in your component
2. **Add stable IDs/classes** to those sections
3. **Implement TargetRenderer** interface
4. **Mark targets dirty** in action handlers
5. **Test** with browser DevTools Network tab

### Example Migration

**Before:**
```go
func (c *Component) Handle(ctx context.Context, action string, data url.Values) error {
    c.Count++
    return nil
}
```

**After:**
```go
func (c *Component) Handle(ctx context.Context, action string, data url.Values) error {
    c.Count++
    c.MarkTargetDirty("#counter") // Add this
    return nil
}

func (c *Component) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
    if c.IsDirty("#counter") {
        return []liveflux.TargetFragment{{
            Selector: "#counter",
            Content:  hb.Span().ID("counter").Text(fmt.Sprintf("%d", c.Count)),
            SwapMode: liveflux.SwapReplace,
        }}
    }
    return nil
}
```

## See Also

- [Proposal](proposals/data-flux-target.md) - Full specification
- [Target Example](../examples/target/) - Working implementation
- [Architecture](architecture.md) - System overview
