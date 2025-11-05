# `data-flux-select`: Response Fragment Filtering

## Overview

This proposal introduces `data-flux-select`, a declarative way for Liveflux clients to extract specific fragments from full HTML responses before applying swaps. The goal is to complement `data-flux-target` (which locates *where* to insert content) with a symmetrical mechanism that chooses *what* to keep from multi-fragment responses, mirroring htmx’s `hx-select` semantics while fitting Liveflux’s component lifecycle.

## Motivation

- **Reduce server branching**: Allow components to return a single comprehensive render while clients pull only the needed fragment for a particular action.
- **Improve compatibility**: Make it easier to integrate with existing controller endpoints that return full pages or legacy templates.
- **Shared payloads**: Support multiple triggers consuming different slices of the same response without duplicating requests.
- **Progressive enhancement**: Preserve full-page responses when JavaScript fails, while providing fine-grained updates when Liveflux is active.

## Current Limitations

1. Liveflux clients expect either fragment templates or a complete component render; there is no filtering step akin to `hx-select` @docs/proposals/partial-updates.md#30-153
2. Components returning full HTML must generate separate payloads for partial consumers to avoid over-updating the DOM
3. External endpoints (e.g., CMS pages) cannot be re-used without server-side customization to emit Liveflux templates

## Proposed Solution

Add a `data-flux-select` attribute that can be placed on action triggers (buttons, links, forms) to declare the CSS selector(s) that should be extracted from an incoming HTML response before swap logic runs.

### Basic Usage

```html
<button data-flux-action="details" data-flux-select="#product-details">
  Show Details
</button>
```

- The server returns a full component (or page) HTML document
- The client parses the response, queries `#product-details`, and only that node is passed to the `data-flux-target`/`data-flux-swap` pipeline

### Multiple Selectors

```html
<button
  data-flux-action="update"
  data-flux-select="#cart-summary, .promo-banner"
  data-flux-target="#cart-summary"
  data-flux-swap="replace"
  data-flux-target-secondary=".promo-banner">
  Refresh Cart
</button>
```

- The response may include both cart and promo markup
- Each selector drives a corresponding target (see below for mapping mechanics)

## Design Details

### Attribute Semantics

- `data-flux-select`: comma-separated CSS selectors to extract from the response document
- Optional `data-flux-select-mode`: `first` (default) selects the first match per selector; `all` collects every match
- Extraction occurs **before** template handling; the filtered nodes become synthetic `<template data-flux-target="…">` entries for the swap pipeline

### Selector-to-Target Mapping

1. **Implicit mapping**: If a selector also appears in `data-flux-target`, the extracted fragment is paired with that target
2. **Secondary targets**: Introduce `data-flux-target:n` (or `data-flux-target-secondary`) to map multiple selectors to distinct targets
3. **Fallback**: If no target is provided, default to the component root (`data-flux-component`), mirroring full render replacement

### Response Handling Flow

1. Client receives raw HTML (string or DOM document)
2. If `data-flux-select` is set on the triggering element:
   - Parse response into a detached DOM
   - For each selector, query matches (respecting `select-mode`)
   - Wrap each match in an in-memory `<template>` element, set `data-flux-target`/`data-flux-swap` based on trigger attributes
3. Pass synthesized templates into `liveflux.applyTargets`
4. If selectors fail to match, fall back to original response (either as a full render or existing templates)

### Interaction with Existing Features

- **`data-flux-target`**: Continues to point at destination nodes; select provides the source fragment
- **`data-flux-swap`**: Still controls how fragments are merged (replace, append, etc.)
- **`data-flux-component` template**: If present in response, it remains available; select-filtered fragments can coexist with server-provided templates
- **WebSocket updates**: When WS messages include full HTML payloads, the select logic applies identically

## Implementation Plan

### Phase 1: Client Support (Weeks 1-2)

1. Extend action metadata resolution to capture `data-flux-select` and optional `data-flux-select-mode`
2. Implement response filtering utility `extractSelectedFragments(responseHTML, selectors, mode)` returning synthetic templates
3. Integrate extraction into HTTP and WebSocket pipelines before `applyTargets`
4. Add console warnings when selectors match nothing; fall back to unfiltered payload
5. Unit tests covering single selector, multiple selectors, `select-mode` variations, and fallback paths

### Phase 2: Trigger-to-Target Mapping (Weeks 3-4)

1. Support indexed `data-flux-target-n` attributes to map selectors → targets
2. Allow specifying swap modes per selector (`data-flux-swap-n`)
3. Integration tests ensuring extracted fragments swap into intended destinations

### Phase 3: Server & DX Enhancements (Weeks 5-6)

1. Document best practices for components returning full HTML when select is used
2. Provide helper constants (`DataFluxSelect`, `DataFluxSelectMode`) and builder APIs in Go helpers @functions.go
3. Update examples to showcase reusing existing templates (e.g., marketing banners)

### Phase 4: Advanced Modes (Weeks 7-8)

1. Introduce select presets (e.g., `data-flux-select="@body"` for common regions)
2. Optional streaming support: allow select to run incrementally if the response is streamed (future)
3. Performance benchmarking to ensure parsing overhead is minimal

## Security & Safety Considerations

- **Out-of-scope extraction**: Warn if selectors reference nodes outside the response’s `<body>`
- **Script execution**: Ensure extracted fragments still pass through `executeScripts` once injected
- **Fallback behavior**: If selectors fail, document that clients revert to full payload to avoid blank updates

## Open Questions

1. Should we allow XPath or only CSS selectors?
2. Should select support negative filters ("keep everything except …")?
3. How do we expose selection results for debugging (e.g., devtools overlay)?
4. Do we need server hints to indicate safe regions (for security/containment)?

## References

- **Partial update protocol**: @docs/proposals/partial-updates.md
- **htmx `hx-select` docs**: https://htmx.org/attributes/hx-select/
