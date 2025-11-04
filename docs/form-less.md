# Form-less Submission: Design & Implementation

## Overview

This document explores extending Liveflux's data submission model to support **selector-based inclusion** of arbitrary DOM elements, inspired by htmx's `hx-include` attribute. The goal is to enable flexible composition patterns while maintaining backward compatibility with the existing form-based approach.

## Current Architecture

### Client Flow

1. **Component initialization**: `Base.Root()` renders a root element with hidden inputs containing `liveflux_component_type` and `liveflux_component_id` @base.go#86-101
2. **Event delegation**: `liveflux_bootstrap.js` registers global click and submit listeners @js/liveflux_bootstrap.js#12-30
3. **Action handling**: `handleActionClick` locates the action button, finds the nearest component root, and serializes either:
   - The associated `<form>` (if the button has a `form` attribute), or
   - The component root itself @js/liveflux_handlers.js#15-54
4. **Serialization**: `serializeElement` collects all `input`, `select`, and `textarea` descendants with `name` attributes; `data-flux-param-*` attributes on the trigger element are merged as additional parameters @js/liveflux_util.js#22-52

### Current Limitations

- **Scope constraint**: Only fields within the component root or an explicitly linked `<form>` are submitted
- **Composition barriers**: Cannot easily share form fragments across components
- **Wrapper overhead**: Requires wrapping non-form elements in `<form>` tags for data collection
- **Cross-component data**: No mechanism to include inputs from sibling or parent components
- **Initialization dependency**: If Liveflux fails to initialize (script error, load failure, etc.), forms submit normally to their `action` URL, causing confusing behavior and potential errors since the server endpoint expects Liveflux-formatted requests

## Proposed Solution

### Inspiration: htmx's `hx-include`

htmx allows `hx-include="#foo"` to include arbitrary DOM nodes during requests. Adopting a similar pattern in Liveflux enables:

1. **Decoupled triggers**: Submit actions from elements outside the component root
2. **Shared forms**: Compose form fragments across multiple components
3. **Flexible markup**: Collect data from non-form elements without wrapper overhead
4. **Progressive enhancement**: Maintain traditional form fallbacks for no-JS scenarios
5. **Initialization safety**: By using `<div>` containers with `data-flux-include` instead of `<form>` elements, avoid accidental native form submissions when Liveflux fails to initialize—buttons without forms simply do nothing rather than submitting to unexpected endpoints

### Proposed API

#### Basic Usage

```html
<!-- Include a specific element by ID -->
<button data-flux-action="save" data-flux-include="#extra-fields">
  Save
</button>

<!-- Include multiple elements -->
<button data-flux-action="submit" 
        data-flux-include="#section-a, #section-b, .shared-inputs">
  Submit All
</button>

<!-- Exclude specific descendants -->
<button data-flux-action="update" 
        data-flux-include="#form-wrapper"
        data-flux-exclude=".ignore-me">
  Update
</button>
```

#### Server-Side Helpers (Future)

```go
// Generate include attributes from Go
hb.Button().
  Attr(liveflux.DataFluxAction, "save").
  Attr(liveflux.DataFluxInclude, "#extra-fields").
  Text("Save")

// Or via a helper method
component.IncludeButton("save", "Save", "#extra-fields", "#shared-data")
```

## Implementation Design

### Phase 1: Client-Side Changes

#### 1. Attribute Recognition

Add support for new attributes in `liveflux_handlers.js`:

- `data-flux-include` / `flux-include`: CSS selectors for elements to include
- `data-flux-exclude` / `flux-exclude`: CSS selectors for elements to exclude from included scopes

#### 2. Enhanced Serialization Logic

Modify `handleActionClick` to:

```javascript
// Pseudo-code for enhanced serialization
function collectFields(btn, root, assocForm) {
  // 1. Serialize default scope (form or root)
  let fields = assocForm 
    ? liveflux.serializeElement(assocForm) 
    : liveflux.serializeElement(root);
  
  // 2. Process data-flux-include
  const includeAttr = btn.getAttribute('data-flux-include') || 
                      btn.getAttribute('flux-include');
  if (includeAttr) {
    const selectors = includeAttr.split(',').map(s => s.trim());
    selectors.forEach(selector => {
      const elements = document.querySelectorAll(selector);
      if (elements.length === 0) {
        console.warn(`[Liveflux] Include selector "${selector}" matched no elements`);
      }
      elements.forEach(el => {
        const included = liveflux.serializeElement(el);
        Object.assign(fields, included); // Later sources override
      });
    });
  }
  
  // 3. Process data-flux-exclude (remove specific keys)
  const excludeAttr = btn.getAttribute('data-flux-exclude') || 
                      btn.getAttribute('flux-exclude');
  if (excludeAttr) {
    const excludeSelectors = excludeAttr.split(',').map(s => s.trim());
    excludeSelectors.forEach(selector => {
      const elements = document.querySelectorAll(selector);
      elements.forEach(el => {
        const excluded = liveflux.serializeElement(el);
        Object.keys(excluded).forEach(key => delete fields[key]);
      });
    });
  }
  
  return fields;
}
```

#### 3. Component Metadata Resolution

Handle cases where the trigger is outside the component root:

```javascript
// Fallback chain for component metadata
function resolveComponentMetadata(btn) {
  // 1. Try nearest root
  let root = btn.closest(rootSelectorWithFallback);
  if (root) {
    const comp = root.querySelector('input[name="liveflux_component_type"]');
    const id = root.querySelector('input[name="liveflux_component_id"]');
    if (comp && id) return { comp: comp.value, id: id.value, root };
  }
  
  // 2. Try explicit attributes on button
  const explicitComp = btn.getAttribute('data-flux-component-type');
  const explicitId = btn.getAttribute('data-flux-component-id');
  if (explicitComp && explicitId) {
    return { comp: explicitComp, id: explicitId, root: null };
  }
  
  // 3. Try data attributes pointing to root
  const rootId = btn.getAttribute('data-flux-root-id');
  if (rootId) {
    root = document.getElementById(rootId);
    // ... extract from root
  }
  
  return null;
}
```

#### 4. Validation & Developer Experience

- **Console warnings**: Log when selectors match zero elements
- **Duplicate field handling**: Document that later sources override (last-write-wins)
- **Circular inclusion**: Detect and warn if include/exclude creates conflicts
- **Performance**: Cache selector results within a single action invocation

#### 5. Testing Strategy

- **Unit tests**: Test `serializeElement` with various DOM structures
- **Integration tests**: 
  - Button with `data-flux-include` submits merged data
  - `data-flux-exclude` removes specific fields
  - Multiple comma-separated selectors work correctly
  - Warnings appear for invalid selectors
  - Component metadata resolution fallback chain

### Phase 2: Server-Side Changes

#### Handler Compatibility

The HTTP handler already processes URL-encoded form payloads via `r.ParseForm()` and extracts fields from `r.Form` @handler.go#67-94. **No structural changes required** for basic functionality.

#### Optional Enhancements

1. **Constants for new attributes**:
   ```go
   // constants.go
   const (
       DataFluxInclude = "data-flux-include"
       DataFluxExclude = "data-flux-exclude"
   )
   ```

2. **Server-side helpers** (future):
   ```go
   // functions.go
   func IncludeAttr(selectors ...string) hb.Attribute {
       return hb.Attr(DataFluxInclude, strings.Join(selectors, ", "))
   }
   
   func ExcludeAttr(selectors ...string) hb.Attribute {
       return hb.Attr(DataFluxExclude, strings.Join(selectors, ", "))
   }
   ```

3. **Validation middleware**: Optional handler wrapper to validate that included selectors don't leak sensitive data across component boundaries

#### Transport Compatibility

- **HTTP POST/GET**: Works immediately with existing handler @handler.go#67-94
- **WebSocket**: Requires mirroring the same serialization logic in `liveflux_websocket.js` to maintain behavioral parity
- **Content-Type**: Continue using `application/x-www-form-urlencoded` for consistency

### Phase 3: WebSocket Integration

The WebSocket transport must serialize fields identically to HTTP:

1. **Shared serialization**: Extract field collection logic into a reusable function used by both HTTP and WS paths
2. **Message format**: WebSocket action messages should include the same merged field set
3. **Testing**: Verify that WS and HTTP produce identical payloads for the same DOM state

```javascript
// Shared in liveflux_util.js
function collectAllFields(btn, root, assocForm) {
  // ... implementation from Phase 1
}

// Used by both liveflux_handlers.js and liveflux_websocket.js
```

## Security & Safety Considerations

### 1. Cross-Component Data Leakage

**Risk**: Arbitrary selectors allow components to read inputs from other components' DOM trees.

**Mitigation strategies**:

- **Option A - Scoped by default**: Only allow `data-flux-include` to reference descendants of the current component root unless `data-flux-include-global` is explicitly set
- **Option B - Namespace validation**: Warn when selectors match elements outside the component's subtree
- **Option C - Server-side validation**: Add optional middleware to validate that submitted fields match expected schema
- **Recommendation**: Start with Option B (warnings) for flexibility, document security implications, and provide Option C for sensitive applications

### 2. Duplicate Field Names

**Current behavior**: `serializeElement` uses last-write-wins for duplicate names @liveflux_util.js#22-38

**Considerations**:

- **Multi-select handling**: Currently collapses to last selected value; consider supporting arrays (`field[]=value1&field[]=value2`)
- **Merge order**: Document that include order matters: `root → form → include selectors (in order) → button params`
- **Explicit override**: Allow `data-flux-param-*` on button to always win

**Proposed precedence** (lowest to highest):
1. Component root fields
2. Associated form fields
3. Included elements (left to right in selector list)
4. Excluded elements (removed)
5. Button `data-flux-param-*` attributes
6. Button `name`/`value` (if applicable)

### 3. Progressive Enhancement

**Requirement**: Ensure graceful degradation when JavaScript is disabled.

**Strategy**:

- Traditional `<form>` elements continue to work via native browser submission
- `data-flux-include` is purely a JS enhancement—no-JS users rely on form scope
- Document that critical fields should live within `<form>` or component root for accessibility
- Consider adding `<noscript>` guidance in documentation

### 4. Performance Implications

**Concerns**:

- `document.querySelectorAll()` on every action could be expensive for complex selectors
- Multiple includes create O(n×m) serialization cost

**Optimizations**:

- Cache selector results during a single action invocation
- Warn developers about expensive selectors (e.g., `*`, deep combinators)
- Consider limiting the number of include selectors (e.g., max 10)
- Profile and document performance characteristics

### 5. CSRF & Input Validation

**Existing protection**: Handler relies on standard CSRF middleware @handler.go#67-94

**New considerations**:

- Included fields bypass traditional form boundaries—ensure CSRF tokens are still submitted
- Document that `Handle()` methods must validate all inputs regardless of source
- Consider adding field allowlists in component metadata

## Use Cases & Examples

### Example 1: Shared Search Filter

```html
<!-- Shared filter panel -->
<div id="global-filters">
  <input type="text" name="search" placeholder="Search...">
  <select name="category">
    <option value="">All</option>
    <option value="tech">Tech</option>
  </select>
</div>

<!-- Component A -->
<div data-flux-root="1" data-flux-component="product-list">
  <!-- ... component content ... -->
  <button data-flux-action="refresh" data-flux-include="#global-filters">
    Refresh Products
  </button>
</div>

<!-- Component B -->
<div data-flux-root="1" data-flux-component="article-list">
  <!-- ... component content ... -->
  <button data-flux-action="refresh" data-flux-include="#global-filters">
    Refresh Articles
  </button>
</div>
```

### Example 2: Multi-Step Form

```html
<!-- Step 1 -->
<div id="step-1">
  <input type="text" name="first_name">
  <input type="text" name="last_name">
</div>

<!-- Step 2 -->
<div id="step-2">
  <input type="email" name="email">
  <input type="tel" name="phone">
</div>

<!-- Component with submit button -->
<div data-flux-root="1" data-flux-component="registration">
  <button data-flux-action="submit" 
          data-flux-include="#step-1, #step-2">
    Complete Registration
  </button>
</div>
```

### Example 3: Excluding Sensitive Fields

```html
<div id="user-form">
  <input type="text" name="username">
  <input type="password" name="password" class="sensitive">
  <input type="text" name="bio">
</div>

<button data-flux-action="update-profile"
        data-flux-include="#user-form"
        data-flux-exclude=".sensitive">
  Update Profile (without password)
</button>
```

## Implementation Roadmap

### Phase 1: Prototype (Week 1-2)

- [ ] Add `data-flux-include` parsing in `liveflux_handlers.js`
- [ ] Implement basic field merging with last-write-wins
- [ ] Add console warnings for empty selectors
- [ ] Feature flag: `window.liveflux.enableFormlessSubmit = true`
- [ ] Unit tests for serialization logic

### Phase 2: Refinement (Week 3-4)

- [ ] Add `data-flux-exclude` support
- [ ] Implement component metadata fallback chain
- [ ] Add performance optimizations (caching, limits)
- [ ] Integration tests with example components
- [ ] Document security considerations

### Phase 3: Server Integration (Week 5-6)

- [ ] Add constants to `constants.go`
- [ ] Create helper functions in `functions.go`
- [ ] Update examples to demonstrate usage
- [ ] Add validation middleware example
- [ ] Update main documentation

### Phase 4: WebSocket Parity (Week 7-8)

- [ ] Extract shared serialization to `liveflux_util.js`
- [ ] Update `liveflux_websocket.js` to use shared logic
- [ ] Cross-transport integration tests
- [ ] Performance benchmarking

### Phase 5: Polish & Release (Week 9-10)

- [ ] Comprehensive documentation
- [ ] Migration guide for existing applications
- [ ] Blog post with examples
- [ ] Add to comparison docs (vs htmx, Livewire, etc.)
- [ ] Release as opt-in feature in v0.x

## Open Questions

1. **Naming**: Should we use `data-flux-include` or `flux-include` (shorter)? Or both for compatibility?
2. **Scope default**: Should selectors be scoped to component subtree by default, or global with warnings?
3. **Multi-value fields**: Support array notation (`field[]`) or keep simple last-write-wins?
4. **Feature flag**: Should this be opt-in via `ClientOptions` or enabled by default with opt-out?
5. **Server helpers**: Priority for `IncludeAttr()` helpers vs. letting users write attributes directly?
6. **Validation**: Should we provide built-in field allowlist validation or leave to user middleware?

## References

- **htmx `hx-include`**: https://htmx.org/attributes/hx-include/
- **Current serialization**: @liveflux_util.js#22-52
- **Current handlers**: @liveflux_handlers.js#15-94
- **Handler flow**: @handler.go#67-94

## Feedback & Discussion

This is a living document. Please provide feedback on:

- API design and naming
- Security implications and mitigations
- Performance concerns
- Use cases not covered
- Implementation priorities
