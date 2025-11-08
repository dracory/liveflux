# Triggers

Liveflux triggers provide declarative event handling for interactive components without writing custom JavaScript. Using the `data-flux-trigger` attribute, you can bind DOM events to component actions with built-in support for debouncing, throttling, and change detection.

## Table of Contents

- [Quick Start](#quick-start)
- [Syntax](#syntax)
- [Event Types](#event-types)
- [Filters](#filters)
- [Modifiers](#modifiers)
- [Examples](#examples)
- [Server-Side Integration](#server-side-integration)
- [Configuration](#configuration)
- [Best Practices](#best-practices)

## Quick Start

Add `data-flux-trigger` to any element to bind events to actions:

```html
<input 
  type="text" 
  name="query"
  data-flux-trigger="keyup changed delay:300ms"
  data-flux-action="search" />
```

This creates a live search that:
- Fires on `keyup` events
- Only triggers if the value changed
- Debounces requests by 300ms

## Syntax

The `data-flux-trigger` attribute accepts a space-separated list of event names, filters, and modifiers:

```
data-flux-trigger="[event] [filter] [modifier]:[value]"
```

Multiple trigger definitions can be separated by commas:

```html
<input data-flux-trigger="keyup delay:300ms, blur" data-flux-action="validate" />
```

### Components

- **Events**: DOM event names (`click`, `keyup`, `change`, `blur`, etc.)
- **Filters**: Conditional logic (`changed`, `once`, `from:selector`, `not:selector`)
- **Modifiers**: Timing and behavior control (`delay:duration`, `throttle:duration`, `queue:strategy`)

## Event Types

### Default Events

When no event is specified, Liveflux infers a sensible default based on element type:

| Element Type | Default Event |
|--------------|---------------|
| `<input type="text\|search\|email\|url\|tel\|password">` | `keyup changed` |
| `<input type="checkbox\|radio">` | `change` |
| `<select>` | `change` |
| `<textarea>` | `keyup changed` |
| `<button>`, `<a>` | `click` |
| `<form>` | `submit` |

Example using default event:

```html
<!-- Infers 'keyup changed' for text input -->
<input type="text" name="search" data-flux-trigger="delay:300ms" data-flux-action="search" />
```

### Explicit Events

Specify any DOM event name:

```html
<!-- Trigger on blur -->
<input data-flux-trigger="blur" data-flux-action="validate" />

<!-- Trigger on focus -->
<input data-flux-trigger="focus" data-flux-action="loadSuggestions" />

<!-- Trigger on input (fires on every change) -->
<input data-flux-trigger="input delay:500ms" data-flux-action="autosave" />
```

## Filters

Filters add conditional logic to triggers.

### `changed`

Only fires if the serialized value has changed since the last trigger:

```html
<input 
  type="text" 
  name="email"
  data-flux-trigger="keyup changed"
  data-flux-action="checkAvailability" />
```

The `changed` filter:
- Serializes all form fields using `collectAllFields`
- Compares with cached value (JSON string comparison)
- Respects `data-flux-include` and `data-flux-exclude`
- Caches values per element using WeakMap

### `once`

Fires only once, then disables the trigger:

```html
<button data-flux-trigger="click once" data-flux-action="subscribe">
  Subscribe (One-time)
</button>
```

### `from:selector`

Only fires if the event originated from an element matching the selector:

```html
<div data-flux-trigger="click from:.delete-btn" data-flux-action="deleteItem">
  <button class="delete-btn">Delete</button>
  <button class="edit-btn">Edit</button>
</div>
```

### `not:selector`

Only fires if the event did NOT originate from an element matching the selector:

```html
<div data-flux-trigger="click not:button" data-flux-action="selectRow">
  <span>Row content</span>
  <button>Action</button>
</div>
```

## Modifiers

Modifiers control timing and request behavior.

### `delay:duration`

Debounces the trigger by the specified duration:

```html
<!-- Wait 300ms after last keystroke -->
<input data-flux-trigger="keyup delay:300ms" data-flux-action="search" />

<!-- Wait 1 second -->
<input data-flux-trigger="input delay:1s" data-flux-action="autosave" />
```

Duration formats:
- `300ms` - Milliseconds
- `1s` - Seconds (converted to milliseconds)
- `0.5s` - Fractional seconds

### `throttle:duration`

Throttles the trigger to fire at most once per duration:

```html
<!-- Fire at most once per second -->
<div data-flux-trigger="scroll throttle:1s" data-flux-action="loadMore">
  <!-- Scrollable content -->
</div>
```

**Difference between `delay` and `throttle`:**
- `delay` (debounce): Waits for a pause in events
- `throttle`: Fires at regular intervals while events occur

### `queue:strategy`

Controls how multiple pending requests are handled:

```html
<!-- Cancel pending request (default) -->
<input data-flux-trigger="keyup delay:300ms queue:replace" data-flux-action="search" />

<!-- Allow concurrent requests -->
<button data-flux-trigger="click queue:all" data-flux-action="process">
  Process
</button>
```

Queue strategies:
- `replace` (default): Cancels any pending request before starting a new one
- `all`: Allows concurrent requests

## Examples

### Live Search

```html
<input 
  type="search" 
  name="query"
  placeholder="Search..."
  data-flux-trigger="keyup changed delay:300ms"
  data-flux-action="search" />
```

### Auto-save Form

```html
<form>
  <input 
    type="text" 
    name="title"
    data-flux-trigger="input delay:1s"
    data-flux-action="autosave" />
  
  <textarea 
    name="content"
    data-flux-trigger="input delay:1s"
    data-flux-action="autosave"></textarea>
</form>
```

### Field Validation

```html
<input 
  type="email" 
  name="email"
  data-flux-trigger="blur"
  data-flux-action="validateEmail" />
```

### Select-Driven Filters

```html
<select 
  name="category"
  data-flux-trigger="change"
  data-flux-action="filterResults">
  <option value="">All Categories</option>
  <option value="electronics">Electronics</option>
  <option value="books">Books</option>
</select>
```

### Infinite Scroll

```html
<div 
  data-flux-trigger="scroll throttle:500ms"
  data-flux-action="loadMore"
  style="height: 400px; overflow-y: auto;">
  <!-- Content -->
</div>
```

### Multiple Triggers

```html
<!-- Validate on blur, auto-save on input -->
<input 
  type="text" 
  name="username"
  data-flux-trigger="blur, input delay:2s"
  data-flux-action="validate" />
```

## Server-Side Integration

### Detecting Trigger Requests

Triggers send the `X-Liveflux-Trigger` header with the event name:

```go
func (c *MyComponent) Handle(ctx context.Context, action string, data url.Values) error {
    // Check if request came from a trigger
    if r, ok := ctx.Value("request").(*http.Request); ok {
        triggerEvent := r.Header.Get("X-Liveflux-Trigger")
        if triggerEvent != "" {
            log.Printf("Triggered by: %s", triggerEvent)
        }
    }
    
    // Handle action normally
    return nil
}
```

### Event-Specific Logic

Implement different validation strategies based on trigger type:

```go
func (c *FormComponent) Handle(ctx context.Context, action string, data url.Values) error {
    if action != "validate" {
        return nil
    }
    
    r := ctx.Value("request").(*http.Request)
    triggerEvent := r.Header.Get("X-Liveflux-Trigger")
    
    switch triggerEvent {
    case "keyup":
        // Light validation during typing
        c.validateFormat()
    case "blur":
        // Full validation on blur
        c.validateFormat()
        c.validateUniqueness()
    default:
        // Full validation for explicit actions
        c.validateAll()
    }
    
    return nil
}
```

### Rate Limiting

Use the trigger header for intelligent rate limiting:

```go
func (c *SearchComponent) Handle(ctx context.Context, action string, data url.Values) error {
    r := ctx.Value("request").(*http.Request)
    
    // More lenient rate limiting for triggers
    if r.Header.Get("X-Liveflux-Trigger") != "" {
        // Allow more frequent requests from triggers
        c.rateLimiter.AllowBurst(10)
    } else {
        // Stricter limits for explicit actions
        c.rateLimiter.AllowBurst(3)
    }
    
    // Process search...
    return nil
}
```

## Configuration

### Global Configuration

Configure trigger behavior globally:

```javascript
// In your page initialization
liveflux.configureTriggers({
  defaultTriggerDelay: 200,  // Minimum delay for all triggers (ms)
  enableTriggers: true        // Enable/disable trigger system
});
```

### Disabling Triggers

Temporarily disable triggers for debugging:

```javascript
liveflux.configureTriggers({ enableTriggers: false });
```

## Best Practices

### 1. Always Use `changed` with Text Inputs

Prevent redundant requests when users navigate with arrow keys:

```html
<!-- Good -->
<input data-flux-trigger="keyup changed delay:300ms" />

<!-- Avoid: fires even when value doesn't change -->
<input data-flux-trigger="keyup delay:300ms" />
```

### 2. Choose Appropriate Delays

- **Search**: 300-500ms (balance responsiveness and request volume)
- **Auto-save**: 1-2s (avoid excessive saves)
- **Validation**: Use `blur` instead of delay

### 3. Use `blur` for Validation

Validate on blur to avoid interrupting user input:

```html
<input 
  type="email" 
  name="email"
  data-flux-trigger="blur"
  data-flux-action="validate" />
```

### 4. Combine with Targeted Updates

Use triggers with `data-flux-target` for efficient partial updates:

```go
func (c *SearchComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
    if c.IsDirty("results") {
        return []liveflux.TargetFragment{
            {
                Selector: "#search-results",
                Content:  c.renderResults(),
            },
        }
    }
    return nil
}
```

### 5. Progressive Enhancement

Ensure forms work without JavaScript:

```html
<form method="POST" action="/search">
  <!-- Works without JS via form submission -->
  <input 
    type="search" 
    name="query"
    data-flux-trigger="keyup changed delay:300ms"
    data-flux-action="search" />
  
  <!-- Fallback submit button -->
  <button type="submit">Search</button>
</form>
```

### 6. Accessibility

Add ARIA attributes for screen readers:

```html
<div role="search">
  <input 
    type="search"
    aria-label="Search"
    data-flux-trigger="keyup changed delay:300ms"
    data-flux-action="search" />
  
  <div 
    id="results"
    role="region" 
    aria-live="polite"
    aria-atomic="true">
    <!-- Results -->
  </div>
</div>
```

### 7. Avoid Trigger Loops

Don't trigger actions that update the triggering element's value:

```html
<!-- Avoid: can cause infinite loops -->
<input 
  name="price"
  data-flux-trigger="input"
  data-flux-action="formatPrice" />
```

Instead, format on blur or use client-side formatting.

## Troubleshooting

### Triggers Not Firing

1. **Check console for errors**: Look for `[Liveflux Triggers]` messages
2. **Verify component metadata**: Ensure element is inside a `data-flux-root`
3. **Check action exists**: Verify `data-flux-action` attribute is present
4. **Inspect trigger parsing**: Use browser dev tools to check if triggers are registered

### Multiple Requests

1. **Add `changed` filter**: Prevents duplicate requests for same value
2. **Increase delay**: Give users more time before triggering
3. **Use `queue:replace`**: Cancel pending requests (default behavior)

### Performance Issues

1. **Increase debounce delay**: Reduce request frequency
2. **Use targeted updates**: Update only changed fragments
3. **Implement server-side caching**: Cache expensive operations
4. **Add rate limiting**: Protect against excessive requests

## Migration Guide

### From Custom JavaScript

**Before:**
```javascript
let timeout;
document.querySelector('#search').addEventListener('keyup', (e) => {
  clearTimeout(timeout);
  timeout = setTimeout(() => {
    // Make request...
  }, 300);
});
```

**After:**
```html
<input 
  id="search"
  data-flux-trigger="keyup changed delay:300ms"
  data-flux-action="search" />
```

### From htmx

Liveflux triggers are inspired by htmx's `hx-trigger`:

```html
<!-- htmx -->
<input hx-trigger="keyup changed delay:300ms" hx-post="/search" />

<!-- Liveflux -->
<input data-flux-trigger="keyup changed delay:300ms" data-flux-action="search" />
```

Key differences:
- Liveflux uses component-based architecture
- Actions are handled by component methods
- Automatic state management and persistence

## See Also

- [Components](components.md) - Component lifecycle and structure
- [Targeted Updates](targeted_updates.md) - Efficient partial rendering
- [Handler and Transport](handler_and_transport.md) - Request handling
- [Examples](../examples/triggers/) - Working code examples
