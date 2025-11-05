# `data-flux-select`: Response Fragment Filtering

## Status

**Priority**: First implementation - standalone feature  
**Dependencies**: None - works with current architecture  
**Related**: Independent of other proposals

## Overview

This proposal introduces `data-flux-select`, a declarative way for Liveflux clients to extract specific fragments from full HTML responses before applying swaps. The goal is to let clients pick *what* to keep from multi-fragment responses while the existing swap pipeline decides *where* to place content, mirroring htmx's `hx-select` semantics but staying compatible with Liveflux's current full-component replacement model.

## Motivation

- **Reduce server branching**: Allow components to return a single comprehensive render while clients pull only the needed fragment for a particular action.
- **Improve compatibility**: Make it easier to integrate with existing controller endpoints that return full pages or legacy templates.
- **Shared payloads**: Support multiple triggers consuming different slices of the same response without duplicating requests.
- **Progressive enhancement**: Preserve full-page responses when JavaScript fails, while providing fine-grained updates when Liveflux is active.

## Current Limitations

1. Liveflux clients expect a complete component render; there is no filtering step to extract specific fragments from the response
2. Components returning full HTML must generate separate payloads for partial consumers to avoid over-updating the DOM
3. External endpoints (e.g., CMS pages) cannot be re-used without server-side customization to emit Liveflux-compatible HTML
4. Current architecture replaces entire component root via `metadata.root.replaceWith(newNode)` @js/liveflux_handlers.js#55-72

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
  data-flux-select="#cart-summary, .promo-banner">
  Refresh Cart
</button>
```

- The response may include both cart and promo markup
- The client will extract both nodes and, in the current implementation, treat the first match as the replacement fragment (narrow selectors with pseudo-classes or `:first-of-type` when you only need one)

## Design Details

### Attribute Semantics

- `data-flux-select`: comma-separated CSS selectors to extract from the response document
- Extraction occurs **before** swap handling; the filtered nodes are treated as standalone fragments that feed directly into the existing replacement pipeline

### Selector-to-Root Mapping

1. **Current behavior**: The first matched fragment replaces the component root (or form submit scope) using existing `root.replaceWith(newNode)` logic @js/liveflux_handlers.js#61
2. **Multiple selectors**: When multiple selectors match, only the first match is used; use CSS pseudo-classes (`:first-of-type`, `:nth-child()`) to select a single node
3. **Future compatibility**: This feature is designed to work seamlessly with future targeted update proposals without modification

### Response Handling Flow

1. Client receives raw HTML (string or DOM document)
2. If `data-flux-select` is set on the triggering element:
   - Parse response into a detached DOM using `DOMParser`
   - For each selector in comma-separated list, attempt to match via `querySelector`
   - Return the `outerHTML` of the first successful match
3. Pass extracted fragment (or original HTML if no match) into existing replacement logic
4. If selectors fail to match, fall back to original response with console warning

### Interaction with Existing Features

- **Component replacement**: Extracted fragments replace the component root using existing `root.replaceWith(newNode)` logic @js/liveflux_handlers.js#61, #108
- **Script execution**: Extracted fragments pass through `liveflux.executeScripts(newNode)` as normal @js/liveflux_handlers.js#62
- **WebSocket updates**: When WS messages include full HTML payloads, the select logic applies identically
- **Form submissions**: Works with both action clicks and form submits without modification
- **Future compatibility**: Designed to integrate with future targeted update features without breaking changes

## Implementation Plan

### Phase 1: Core Functionality

1. **Attribute capture**: Extend action metadata resolution to capture `data-flux-select` from trigger elements (buttons, forms)
2. **Extraction utility**: Implement `liveflux.extractSelectedFragments(responseHTML, selectors)` that:
   - Parses response HTML into a detached DOM
   - Queries each selector and returns the first match
   - Returns original HTML if no matches found
3. **Integration points**:
   - Modify `handleActionClick` to apply extraction before `tmp.innerHTML = html` @js/liveflux_handlers.js#57-58
   - Modify `handleFormSubmit` to apply extraction before `tmp.innerHTML = html` @js/liveflux_handlers.js#104-105
   - Modify WebSocket `handleUpdate` to apply extraction before DOM replacement
4. **Fallback behavior**: Log console warning when selectors fail; use unfiltered response
5. **Testing**: Unit tests for single selector, multiple selectors, no-match fallback, and malformed selectors

### Phase 2: Developer Experience

1. Add debug logging toggle for selection results
2. Document selector best practices and common patterns
3. Update examples to showcase fragment extraction use cases

## Security & Safety Considerations

- **Out-of-scope extraction**: Warn if selectors reference nodes outside the response’s `<body>`
- **Script execution**: Ensure extracted fragments still pass through `executeScripts` once injected
- **Fallback behavior**: If selectors fail, document that clients revert to full payload to avoid blank updates

## Open Questions

1. Should we allow XPath or only CSS selectors? **Recommendation**: Start with CSS selectors only for simplicity
2. Should select support negative filters ("keep everything except …")? **Recommendation**: Defer to future iteration
3. How do we expose selection results for debugging? **Recommendation**: Console logging with optional verbose mode
4. Do we need server hints to indicate safe regions? **Recommendation**: Not required; client-side selection is sufficient
5. Should we support multiple fragments or only first match? **Recommendation**: First match only initially; multiple fragment support can be added later if needed

## References

- **htmx `hx-select` docs**: https://htmx.org/attributes/hx-select/
- **Action handlers**: @js/liveflux_handlers.js#18-121
- **WebSocket updates**: @js/liveflux_websocket.js
- **Component rendering**: @handler.go#254-269

## Example Implementation Snippet

```javascript
// Add to liveflux namespace
liveflux.extractSelectedFragments = function(html, selectors) {
  if (!selectors || selectors.trim() === '') {
    return html;
  }
  
  const parser = new DOMParser();
  const doc = parser.parseFromString(html, 'text/html');
  const selectorList = selectors.split(',').map(s => s.trim());
  
  for (const selector of selectorList) {
    try {
      const match = doc.querySelector(selector);
      if (match) {
        console.log('[Liveflux Select] Extracted fragment:', selector);
        return match.outerHTML;
      }
    } catch (e) {
      console.warn('[Liveflux Select] Invalid selector:', selector, e);
    }
  }
  
  console.warn('[Liveflux Select] No matches found for selectors:', selectors);
  return html; // Fallback to original
};

// Usage in handleActionClick (before line 57)
const selectAttr = btn.getAttribute('data-flux-select');
const html = selectAttr 
  ? liveflux.extractSelectedFragments(result.html || result, selectAttr)
  : (result.html || result);
```
