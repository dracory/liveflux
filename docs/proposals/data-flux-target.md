# `data-flux-target`: Targeted Fragment Updates

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

- `data-flux-swap="replace"` (default) – replace the matched node’s `outerHTML`
- `data-flux-swap="inner"` – replace the node’s `innerHTML`
- `data-flux-swap="beforebegin" | "afterend"` – insert the fragment as a sibling
- `data-flux-swap="beforebegin-keep" | "afterend-keep"` – insert fragment while leaving marker in place for streaming lists (future)
- Additional strategies (morphing, prepend/append) can layer on this mechanism in later phases

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

1. **Handshake flag**: Extend `liveflux.post` (and form submits) to send `X-Liveflux-Target: only` when the component declares target support
2. **Fragment parsing**: Parse the response into a DOM document and collect `<template data-flux-target>` nodes
3. **Patch application**: Add `liveflux.applyTargets(templates)` to iterate templates, resolve selectors via `document.querySelectorAll(template.dataset.fluxTarget)`, honor `data-flux-swap` semantics, enforce component metadata (skip if `data-flux-component(-id)` doesn’t match the target node’s owning component), perform the appropriate DOM mutation for each match, and re-run `executeScripts` for any new nodes. If a `<template data-flux-component>` is present, replace the existing component root before applying other selectors to ensure the component stays canonical.
4. **Resync pathway**: Use the same handler without the handshake flag to serve a full render when clients need to recover
5. **No-op default**: Components not implementing partial rendering continue full-root behavior

### Phase 2: Server Rendering Hooks

1. **Dirty tracking**: Extend `Base` to track target keys marked via `MarkTargetDirty("cart-total")`
2. **Render API**: Introduce optional `TargetRenderer` interface with `RenderTargets(ctx context.Context) map[string]hb.Node`
3. **Handler integration**: When the handshake flag is present, stream fragment templates produced by `RenderTargets`, including desired `data-flux-swap` values and optional `data-flux-component(-id)` metadata, and optionally attach a `<template data-flux-component>` containing the full render; otherwise render the full component as today @handler.go#201-203
4. **Resync pathway**: Use the same handler without the handshake flag to serve a full render when clients need to recover
5. **No-op default**: Components not implementing partial rendering continue full-root behavior

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

## Open Questions

1. **Selector Strategy**: Should we require `data-flux-partial` markers, or allow arbitrary CSS selectors provided by components?
2. **Dirty Tracking API**: How should components declare state changes—manual `MarkPartialDirty`, automatic diffing, or template annotations?
3. **DOM Morphing**: Should we introduce a morphing library (e.g., `morphdom`) to reduce replacements when structure is similar?
4. **Streaming Support**: Do we need to support incremental partial streams (SSE/WS) before the main response completes?
5. **Testing Utilities**: Provide test helpers to assert partial payloads in component tests?

## References

- **Action handlers**: @js/liveflux_handlers.js#18-111
- **Bootstrap flow**: @js/liveflux_bootstrap.js#12-30
- **WebSocket updates**: @js/liveflux_websocket.js#70-89
- **Server rendering**: @handler.go#161-270
