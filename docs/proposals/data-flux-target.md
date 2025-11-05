# `data-flux-target`: Targeted Fragment Updates

## Status

**Priority**: Second implementation - builds on `data-flux-select`  
**Dependencies**: None (can be implemented independently, but complements `data-flux-select`)  
**Complexity**: High - requires both client and server changes  
**Timeline**: 8-10 weeks for full implementation

## Overview

This proposal introduces **targeted fragment updates** so Liveflux can replace only specific DOM regions instead of re-rendering an entire component root after every interaction. The goal is to reduce network payloads, preserve client-side state (focus, scroll, video playback), and enable more granular UI composition while keeping backward compatibility with the current full-root rendering model.

## Current Architecture

### Client Flow

1. **Action dispatch**: `handleActionClick` and `handleFormSubmit` gather parameters, post them, and replace the component root with returned HTML @js/liveflux_handlers.js#18-111
2. **Placeholder bootstrap**: Components mount via `liveflux.mountPlaceholders()` and global listeners established in `liveflux_bootstrap.js` @js/liveflux_bootstrap.js#12-30
3. **WebSocket updates**: When WebSocket transport is active, `handleUpdate` replaces the root element’s `outerHTML` with the received markup, then reinitializes listeners @js/liveflux_websocket.js#70-89

### Server Flow

1. **Action handling**: `Handler.handle` processes actions, persists component state, and always re-renders the entire component via `writeRender` @handler.go#161-270
2. **Rendering output**: `c.Render(ctx).ToHTML()` produces the full component tree; no metadata about subtrees or fragments is returned @handler.go#254-269

### Current Limitations

- **Whole-root replacement**: Any interaction swaps the entire DOM subtree, causing focus loss, scroll jumps, and widget reset
- **Network overhead**: Large components send full HTML even when only a small portion changed
- **Inefficient animations**: CSS transitions restart because nodes are replaced, not mutated
- **WebSocket parity**: WS updates mirror HTTP behavior; there is no incremental diffing API
- **Stateful integrations**: Third-party scripts attached to subtrees must re-run after every update

## Proposed Solution

Introduce a target-based update protocol where the server may return only the minimal fragments required to update dirty regions. Clients opt in via a "targets-only" handshake flag, apply fragments in place using CSS selectors or stable identifiers, and request a one-time full render if they detect a mismatch. A companion `data-flux-swap` attribute mirrors htmx-style swap semantics (e.g., replace vs. append) so components control how fragments are merged.

### Target Update Payload

- **Handshake**: Client adds a header or form field (e.g., `X-Liveflux-Target: only`) to declare support for fragment-only responses
- **Response body**: Server returns a fragment container document composed of zero or more `<template data-flux-target="…">` nodes plus an optional `<template data-flux-component>` block containing a full component render. Templates **may** include `data-flux-component` / `data-flux-component-id` metadata when a fragment should only apply to a specific component instance; otherwise the selector alone determines the target.
- **Swap mode**: Each template may specify `data-flux-swap` (`replace`, `inner`, `beforebegin`, `afterbegin`, `beforeend`, `afterend`; default `replace`) to control how the fragment is merged with the matched node
- **Resync**: If the client fails to match any selector, it retries the request without the handshake flag to fetch a traditional full render once

```html
<!-- Targeted fragments (component metadata optional) -->
<template
  data-flux-target="#cart-total"
  data-flux-swap="replace">
  <span id="cart-total">$125</span>
</template>

<template
  data-flux-target=".line-items"
  data-flux-swap="beforeend"
  data-flux-component="cart"
  data-flux-component-id="abc123">
  <li class="line-item" data-id="42">…</li>
</template>

<!-- Optional full component render -->
<template data-flux-component="cart" data-flux-component-id="abc123" data-flux-swap="replace">
  <div data-flux-root="1" data-flux-component="cart" data-flux-component-id="abc123">
    <!-- full HTML render -->
  </div>
</template>
```

- **Selector**: Derived from the literal CSS selector stored in `data-flux-target`; helpers can still emit selectors based on component metadata
- **Ordering**: Apply templates in document order; later templates overwrite earlier ones on conflict

### Proposed API

#### Example DOM with Stable Selectors

```html
<div data-flux-root="1" data-flux-component="cart" data-flux-component-id="abc123">
  <span id="cart-total">$99</span>
  <ul class="line-items">
    <li class="line-item" data-id="41">...</li>
  </ul>
</div>
```

#### Response Format (Fragments + Optional Component)

```html
<template
  data-flux-target="#cart-total"
  data-flux-swap="replace"
  data-flux-component="cart"
  data-flux-component-id="abc123">
  <span id="cart-total">$125</span>
</template>

<template
  data-flux-target=".line-items"
  data-flux-swap="beforeend"
  data-flux-component="cart"
  data-flux-component-id="abc123">
  <li class="line-item" data-id="42">…</li>
</template>

<template data-flux-component data-flux-component="cart" data-flux-component-id="abc123">
  <div data-flux-root data-flux-component="cart">
    <!-- replaced component root -->
  </div>
</template>
```

#### Swap Controls

- `data-flux-swap="replace"` (default) – replace the matched node's `outerHTML`
- `data-flux-swap="inner"` – replace the node's `innerHTML`
- `data-flux-swap="beforebegin"` – insert fragment before the target element
- `data-flux-swap="afterbegin"` – insert fragment as first child of target
- `data-flux-swap="beforeend"` – insert fragment as last child of target (append)
- `data-flux-swap="afterend"` – insert fragment after the target element
- Additional strategies (morphing, delete) can layer on this mechanism in later phases

**Swap Mode Reference**:
```javascript
// Implementation in liveflux.applyTargets
switch (swapMode) {
  case 'replace': target.replaceWith(fragment); break;
  case 'inner': target.innerHTML = fragment.innerHTML; break;
  case 'beforebegin': target.insertAdjacentElement('beforebegin', fragment); break;
  case 'afterbegin': target.insertAdjacentElement('afterbegin', fragment); break;
  case 'beforeend': target.insertAdjacentElement('beforeend', fragment); break;
  case 'afterend': target.insertAdjacentElement('afterend', fragment); break;
}
```

#### Server-Side Rendering Helpers (Future)

```go
func (c *CartComponent) RenderTargets(ctx context.Context, dirty map[string]bool) map[string]hb.Node {
    if dirty["cart-total"] {
        targets["#cart-total"] = c.renderTotal(ctx)
    }
    if dirty["line-item-42"] {
        targets[".line-items"] = c.renderLineItem(ctx, 42)
    }
    return targets
}
```

Components track dirty regions and generate fragment markup keyed by selectors and desired swap modes. Helper methods can still translate symbolic keys into selectors (`TargetSelector("cart-total") -> "#cart-total"`). Components that require a complete re-render emit both selector fragments (for cross-component updates) and a `<template data-flux-component>` for their own root. When emitting fragments for other components, include `data-flux-component` / `data-flux-component-id` attributes so the client can verify ownership before applying swaps; when omitted, swaps operate purely on selector matches.

## Implementation Design

### Phase 1: Client-Side Infrastructure

1. **Handshake flag**: Extend `liveflux.post` to send `X-Liveflux-Target: enabled` header when client supports targeted updates
2. **Fragment parsing**: Parse response into DOM document and collect `<template data-flux-target>` nodes
3. **Patch application**: Implement `liveflux.applyTargets(templates, componentRoot)` to:
   - Process `<template data-flux-component>` first if present (full component replacement)
   - Iterate remaining templates in document order
   - For each template, resolve selector via `componentRoot.querySelector(template.dataset.fluxTarget)`
   - Validate component metadata matches (skip if `data-flux-component(-id)` doesn’t match)
   - Apply swap mode from `data-flux-swap` attribute (default: `replace`)
   - Execute scripts in new nodes via `liveflux.executeScripts(newNode)` @js/liveflux_handlers.js#62
   - Log warnings for failed selector matches
4. **Integration points**:
   - Modify `handleActionClick` to detect template responses and call `applyTargets` instead of `replaceWith` @js/liveflux_handlers.js#55-72
   - Modify `handleFormSubmit` similarly @js/liveflux_handlers.js#102-111
   - Modify WebSocket `handleUpdate` to support template format @js/liveflux_websocket.js#78-89
5. **Resync pathway**: If all selectors fail, retry request without handshake flag to get full render
6. **Backward compatibility**: Components not using templates continue full-root replacement

### Phase 2: Server Rendering Hooks

1. **Dirty tracking**: Extend `Base` struct to include:
   ```go
   type Base struct {
       // ... existing fields
       dirtyTargets map[string]bool
   }
   
   func (b *Base) MarkTargetDirty(key string) {
       if b.dirtyTargets == nil {
           b.dirtyTargets = make(map[string]bool)
       }
       b.dirtyTargets[key] = true
   }
   ```
2. **Render API**: Introduce optional `TargetRenderer` interface:
   ```go
   type TargetRenderer interface {
       RenderTargets(ctx context.Context) []TargetFragment
   }
   
   type TargetFragment struct {
       Selector  string    // CSS selector
       Content   hb.Node   // HTML content
       SwapMode  string    // replace, inner, beforeend, etc.
   }
   ```
3. **Handler integration**: Modify `handler.go` to:
   - Check for `X-Liveflux-Target: enabled` header
   - If present and component implements `TargetRenderer`, call `RenderTargets()`
   - Build response with `<template>` nodes for each fragment
   - Optionally include `<template data-flux-component>` with full render as fallback
   - Otherwise render full component as today @handler.go#254-269
4. **Response builder**: Add helper to construct template-based responses:
   ```go
   func buildTargetResponse(fragments []TargetFragment, fullRender string, comp ComponentInterface) string
   ```
5. **Backward compatibility**: Components not implementing `TargetRenderer` continue full-root behavior

### Phase 3: Developer Experience

1. **Helper builders**: Add `TargetAttr(selector string)` returning `data-flux-target="selector"`
2. **Swap helpers**: Provide `SwapAttr(mode string)` to set `data-flux-swap` with compile-time-safe values
3. **Component metadata helpers**: Provide `ComponentMetaAttrs(component string, id string)` returning `data-flux-component` / `data-flux-component-id`
4. **Selector helpers**: Support hierarchical keys (e.g., `list:42`) and convert to deterministic selectors (e.g., `TargetSelector("list", 42) -> ".line-items [data-id='42']")
5. **Component template coordination**: Ensure a `<template data-flux-component>` applies before nested selector swaps to avoid double work
6. **Nested updates**: Ensure parent target applying does not re-render children already patched in same cycle (apply deepest-first)

### Phase 4: WebSocket Parity

1. **WS payload schema**: Mirror HTTP fragment containers (templates) and reuse target application logic, including `data-flux-swap` semantics and `<template data-flux-component>` handling @js/liveflux_websocket.js#78-88
2. **Reconnect safety**: On reconnect or initial mount, suppress the handshake so the server sends a full render, then re-enable targets for subsequent messages
3. **Latency tests**: Benchmark swap-mode performance under rapid updates

### Phase 5: Tooling & Documentation

1. **Developer ergonomics**: CLI lint or build helpers verifying selectors resolve to elements during build/tests
2. **Docs**: Tutorial on marking targets, selector hygiene, debugging missing matches, and best practices
3. **Examples**: Update sample apps (cart, dashboard graphs) to showcase selector-driven updates

## Security & Safety Considerations

1. **Selector abuse**: Restrict selectors to within component root unless `data-flux-target-global` is set; warn if matches fall outside @js/liveflux_handlers.js#18-71
2. **Out-of-sync DOM**: Detect failed selector matches; revert to full render to avoid broken UI
3. **CSRF & Validation**: Existing form handling remains unchanged; partial updates do not affect request payloads @handler.go#214-221
4. **Script injection**: Apply the same sanitization as full renders; partial HTML originates from server-controlled templates
5. **Focus management**: Provide optional utilities to persist focus if partial patch replaces currently focused element

## Use Cases & Examples

### Live Cart Totals

Update cart total and badge count without re-rendering entire checkout form.

### Real-Time Dashboards

Stream WebSocket updates for individual charts or KPIs while leaving surrounding layout untouched.

### Form Validation Feedback

Return validation errors for specific fields, replacing only error messages and input wrappers.

### Infinite Lists

Append newly loaded list items as partials without disturbing scrolled content.

## Implementation Roadmap

### Phase 1: Prototype (Weeks 1-2)

- [ ] Extend client response parsing and add partial application helper
- [ ] Implement fallback logic and instrumentation (warnings, metrics)
- [ ] Unit tests for DOM patching

### Phase 2: Server Hooks (Weeks 3-4)

- [ ] Add dirty tracking API and `PartialRenderer` interface
- [ ] Update handler to include partial payloads
- [ ] Integration tests with sample component

### Phase 3: Developer Experience (Weeks 5-6)

- [ ] Introduce template helpers and selector conventions
- [ ] Update component examples and documentation
- [ ] Add debug logging toggles for partial matching

### Phase 4: WebSocket Parity (Weeks 7-8)

- [ ] Share partial application between HTTP and WS paths
- [ ] Stress test under rapid updates
- [ ] Validate reconnection behavior

### Phase 5: Release Preparation (Weeks 9-10)

- [ ] Comprehensive docs & migration guide
- [ ] Feature flag via `ClientOptions` and global JS toggle
- [ ] Blog post / announcement and comparison docs updates

### Phase 6: Advanced Features (Future)

- [ ] DOM morphing integration (e.g., `idiomorph`) for minimal DOM changes
- [ ] Streaming support for incremental updates
- [ ] Automatic dirty tracking via state diffing
- [ ] Visual debugging tools (highlight updated regions)
- [ ] Performance profiling and optimization

## Relationship with `data-flux-select`

While both proposals deal with fragment handling, they serve different purposes:

| Feature | `data-flux-select` | `data-flux-target` |
|---------|-------------------|--------------------|
| **Direction** | Client extracts from server response | Server sends specific fragments |
| **Complexity** | Low (client-only) | High (client + server) |
| **Use Case** | Reuse existing endpoints, filter full HTML | Optimize network, targeted updates |
| **Server Changes** | None required | Requires `TargetRenderer` interface |
| **Swap Control** | No (always replaces root) | Yes (`data-flux-swap` modes) |
| **Multiple Targets** | No (first match only) | Yes (multiple fragments per response) |

**Recommendation**: Implement `data-flux-select` first for quick wins, then add `data-flux-target` for advanced use cases.

## Open Questions

1. **Selector Strategy**: Allow arbitrary CSS selectors; provide helper functions for common patterns. **Decision**: Start with arbitrary selectors, add validation helpers later.
2. **Dirty Tracking API**: Manual `MarkTargetDirty()` initially; consider automatic diffing in future. **Decision**: Manual tracking for explicit control.
3. **DOM Morphing**: Defer morphing to Phase 6; start with simple replacement. **Decision**: Add morphing as opt-in feature later if needed.
4. **Streaming Support**: Not required for MVP; can add SSE/chunked encoding later. **Decision**: Single response initially.
5. **Testing Utilities**: Yes, provide `AssertTargetFragment(t, response, selector, expected)` helper. **Decision**: Include in Phase 3.

## Complete Implementation Example

### Client-Side (JavaScript)

```javascript
// Add to liveflux namespace
liveflux.applyTargets = function(html, componentRoot) {
  const parser = new DOMParser();
  const doc = parser.parseFromString(html, 'text/html');
  const templates = doc.querySelectorAll('template[data-flux-target], template[data-flux-component]');
  
  if (templates.length === 0) {
    // No templates found, treat as full component replacement
    return html;
  }
  
  let appliedCount = 0;
  
  // Process component template first (full replacement)
  const componentTemplate = doc.querySelector('template[data-flux-component]:not([data-flux-target])');
  if (componentTemplate) {
    const newRoot = componentTemplate.content.firstElementChild;
    if (newRoot) {
      componentRoot.replaceWith(newRoot);
      liveflux.executeScripts(newRoot);
      componentRoot = newRoot; // Update reference for subsequent selectors
      appliedCount++;
    }
  }
  
  // Process targeted fragments
  templates.forEach(template => {
    if (template.hasAttribute('data-flux-component') && !template.hasAttribute('data-flux-target')) {
      return; // Already processed above
    }
    
    const selector = template.dataset.fluxTarget;
    const swapMode = template.dataset.fluxSwap || 'replace';
    
    try {
      const target = componentRoot.querySelector(selector);
      if (!target) {
        console.warn('[Liveflux Target] Selector not found:', selector);
        return;
      }
      
      const fragment = template.content.firstElementChild;
      if (!fragment) return;
      
      // Apply swap mode
      switch (swapMode) {
        case 'replace':
          target.replaceWith(fragment);
          liveflux.executeScripts(fragment);
          break;
        case 'inner':
          target.innerHTML = fragment.innerHTML;
          liveflux.executeScripts(target);
          break;
        case 'beforebegin':
        case 'afterbegin':
        case 'beforeend':
        case 'afterend':
          target.insertAdjacentElement(swapMode, fragment);
          liveflux.executeScripts(fragment);
          break;
        default:
          console.warn('[Liveflux Target] Unknown swap mode:', swapMode);
      }
      
      appliedCount++;
      console.log('[Liveflux Target] Applied:', selector, 'mode:', swapMode);
    } catch (e) {
      console.error('[Liveflux Target] Error applying selector:', selector, e);
    }
  });
  
  if (appliedCount === 0) {
    console.warn('[Liveflux Target] No targets applied, falling back to full render');
    return html; // Signal to caller to do full replacement
  }
  
  return null; // Targets applied successfully
};

// Modify handleActionClick to use applyTargets
// In liveflux_handlers.js after line 55:
liveflux.post(params).then((result) => {
  const html = result.html || result;
  
  // Check if response contains templates
  if (html.includes('<template data-flux-target') || html.includes('<template data-flux-component')) {
    const fallback = liveflux.applyTargets(html, metadata.root);
    if (fallback) {
      // Targets failed, do full replacement
      const tmp = document.createElement('div');
      tmp.innerHTML = fallback;
      const newNode = tmp.firstElementChild;
      if (newNode) metadata.root.replaceWith(newNode);
    }
  } else {
    // Traditional full replacement
    const tmp = document.createElement('div');
    tmp.innerHTML = html;
    const newNode = tmp.firstElementChild;
    if (newNode) metadata.root.replaceWith(newNode);
  }
  
  if (liveflux.initWire) liveflux.initWire();
});
```

### Server-Side (Go)

```go
// Add to base.go
type TargetRenderer interface {
    RenderTargets(ctx context.Context) []TargetFragment
}

type TargetFragment struct {
    Selector  string
    Content   string
    SwapMode  string
}

// Modify handler.go writeRender function
func (h *Handler) writeRender(ctx context.Context, w http.ResponseWriter, c ComponentInterface) {
    // Check if client supports targets
    supportsTargets := r.Header.Get("X-Liveflux-Target") == "enabled"
    
    if supportsTargets {
        if tr, ok := c.(TargetRenderer); ok {
            fragments := tr.RenderTargets(ctx)
            if len(fragments) > 0 {
                // Build template response
                var sb strings.Builder
                for _, frag := range fragments {
                    sb.WriteString(fmt.Sprintf(
                        `<template data-flux-target="%s" data-flux-swap="%s">%s</template>`,
                        frag.Selector,
                        frag.SwapMode,
                        frag.Content,
                    ))
                }
                
                // Optionally include full render as fallback
                fullRender := c.Render(ctx).ToHTML()
                sb.WriteString(fmt.Sprintf(
                    `<template data-flux-component="%s" data-flux-component-id="%s">%s</template>`,
                    c.GetAlias(),
                    c.GetID(),
                    fullRender,
                ))
                
                w.Header().Set("Content-Type", "text/html; charset=utf-8")
                w.Write([]byte(sb.String()))
                return
            }
        }
    }
    
    // Fallback to full render
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(c.Render(ctx).ToHTML()))
}

// Example component implementation
type CartComponent struct {
    Base
    Total float64
    Items []CartItem
}

func (c *CartComponent) RenderTargets(ctx context.Context) []TargetFragment {
    // Only render what changed
    return []TargetFragment{
        {
            Selector: "#cart-total",
            Content:  fmt.Sprintf(`<span id="cart-total">$%.2f</span>`, c.Total),
            SwapMode: "replace",
        },
        // Add more fragments as needed
    }
}
```

## References

- **Action handlers**: @js/liveflux_handlers.js#18-121
- **Bootstrap flow**: @js/liveflux_bootstrap.js#12-35
- **WebSocket updates**: @js/liveflux_websocket.js#78-89
- **Server rendering**: @handler.go#254-269
- **Related proposal**: @docs/proposals/data-flux-select.md
