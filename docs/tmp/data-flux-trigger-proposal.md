# Proposal: Declarative Trigger Modifiers for Liveflux

## Summary
Introduce a `data-flux-trigger` attribute that lets authors declare client-side event bindings with optional filters (e.g. `changed`) and timing modifiers (e.g. `delay:300ms`, `throttle:1s`). The goal is to support patterns currently solved with custom JavaScript, such as firing component actions on `keyup` or `change` with debounce, while keeping the Liveflux runtime HTML-first and compatible with targeted fragment updates.

## Motivation
- **Parity with htmx**: teams moving from htmx expect `hx-trigger`-like ergonomics.
- **Livewire lessons**: Livewire’s directive/modifier system shows how first-class event bindings with debounce/throttle improve DX without extra JS.
- **Lower JS footprint**: avoid hand-written listeners for common interactions (search-as-you-type, select-change refreshes, blur validation, etc.).
- **Progressive enhancement**: keep forms functional without JavaScript while enabling richer behavior when the client runtime is loaded.
- **Consistent API surface**: reuse Liveflux attribute naming (`data-flux-*`) and align with existing concepts like `data-flux-select`.

## Goals
1. Declaratively bind DOM events to Liveflux actions without extra scripts.
2. Support event filters (`changed`, `once`, `from:<selector>`) to reduce redundant requests, with value-change caching baked in by default.
3. Provide timing modifiers (`delay`, `throttle`) and queue strategies (`replace`, `all`) to control frequency.
4. Maintain backwards compatibility—existing click and submit flows continue to work.
5. Integrate cleanly with targeted fragment updates and request indicators.

## Out of Scope (Initial Version)
- Custom user-defined modifiers beyond the built-in set.
- Complex gesture or multi-event sequences (e.g. `keyup changed, blur`).
- Streaming responses; focus on standard Liveflux postbacks.
- Advanced queue strategies (`first`, `last`) beyond `replace` and `all`.

## Proposed Authoring API
```html
<input
  name="query"
  data-flux-trigger="keyup changed delay:300ms"
  data-flux-action="search" />
```

Syntax mirrors htmx:
- **Event list**: space-separated entries; commas separate distinct trigger definitions.
- **Filters**: `changed`, `once`, `from:<selector>`, `not:<selector>`.
- **Timing modifiers**: `delay:<duration>`, `throttle:<duration>`, `queue:<strategy>` (default `replace`; initial version supports `replace` and `all` only).
- **Action resolution**: reuses `data-flux-action` on the same element; falls back to the nearest action definition in ancestors.
- **Default events**: omitting an explicit event infers a sensible default based on element type (e.g. inputs → `keyup changed`, selects → `change`).

### Default Behavior
- When `data-flux-trigger` is present, Liveflux no longer requires a click/submit to fire the action.
- If both trigger and native submit exist, the first event to fire owns the request; subsequent triggers respect queuing rules.
- **Default event inference**: When no explicit event is specified (e.g., `data-flux-trigger="delay:300ms"`), the runtime infers:
  - `<input type="text|search|email|url|tel|password">` → `keyup changed`
  - `<input type="checkbox|radio">`, `<select>` → `change`
  - `<textarea>` → `keyup changed`
  - `<button>`, `<a>` → `click`
  - `<form>` → `submit`
  - Authors can override with explicit event names: `data-flux-trigger="blur delay:300ms"`

### Form Integration
- For form controls, the runtime serializes fields using the existing `collectAllFields` helper so `data-flux-include` / `data-flux-exclude` continue to work.
- For standalone elements, runtime locates component metadata via `resolveComponentMetadata`, identical to click handling.

## Design Influences (Livewire)
- **Directive-style ergonomics**: mirror Livewire’s `wire:model.debounce` by keeping modifiers declarative and close to the markup.
- **Built-in modifiers**: ship `delay`, `throttle`, and key filters out of the box so common debounce/throttle cases need no custom JS.
- **Value-change awareness**: cache last serialized values per element/root to make `changed` semantics automatic and efficient.
- **Lifecycle resilience**: re-bind triggers after targeted updates, just as Livewire re-initializes directives after DOM morphing.
- **Server awareness**: reuse component metadata and wire helpers so actions submit with the right context every time.

## Shortcut Attributes (Phase 2)
To match the ergonomics of Livewire (`wire:*`) and LiveView (`phx-*`), we may ship syntactic sugar for common patterns **after validating the core API**. These shortcuts reduce duplication when the trigger + action are tightly coupled.

| Shorthand Attribute | Implied Trigger | Notes |
| ------------------- | --------------- | ----- |
| `data-flux-change="update"` | Equivalent to `data-flux-action="update" data-flux-trigger="change changed"` | Includes `changed` guard automatically. |
| `data-flux-keyup="search"` | Equivalent to `data-flux-action="search" data-flux-trigger="keyup changed"` | Pairs with modifiers like `delay:300ms`. |
| `data-flux-blur="validate"` | Equivalent to `data-flux-action="validate" data-flux-trigger="blur"` | Common for field validation. |

**Note**: We intentionally exclude `data-flux-click` and `data-flux-submit` shortcuts to avoid confusion with existing `data-flux-action` semantics. For default click/submit behavior, continue using `data-flux-action`. Shortcuts are reserved for non-default event bindings.

Shortcuts can still accept modifiers via an additional `data-flux-trigger-modifiers` attribute (e.g., `data-flux-keyup="search" data-flux-trigger-modifiers="delay:300ms"`) or by overriding with an explicit `data-flux-trigger` when more control is needed. The runtime normalizes all shorthands into full trigger definitions during parsing, so downstream systems treat them identically.

## Runtime Changes
1. **Parser**: add `liveflux.parseTriggers(el)` returning an array of `{ events, filters, modifiers }`, including inferred default events when authors omit them.
2. **Listener registry**: maintain per-element listeners; ensure cleanup when DOM fragments are replaced (tie into `executeScripts`, `applyTargets`, and `initWire`).
3. **Event handling**:
   - Evaluate filters (`changed` compares serialized value cache; `from`/`not` inspect `event.target`; key filters match `event.key`).
   - Apply timing modifiers using debounced / throttled wrappers (store per trigger state) similar to Livewire's `.debounce` / `.throttle`.
   - Respect `queue` strategies: `replace` (default, cancels pending request), `all` (allows concurrent requests). Future versions may add `first` and `last`.
   - Persist last-known values per element/form so `changed` can short-circuit redundant requests. Values are serialized using `collectAllFields` and compared as JSON strings. Cache is scoped per element and respects `data-flux-include`/`data-flux-exclude`.
4. **Request dispatch**: reuse `post` pipeline, including targeted templates, indicators, and error handling.
5. **Global bootstrap**: on `bootstrapInit`, scan for `[data-flux-trigger]` elements and register listeners. Re-run after any DOM update (`applyTargets`, `executeScripts`, `initWire`) to catch newly-inserted nodes.

## Server Considerations
- **Protocol enhancement**: Add optional `X-Liveflux-Trigger` header containing the triggering event name (e.g., `keyup`, `change`) to help handlers distinguish trigger-driven requests from explicit actions. Useful for rate limiting or different validation strategies.
- Standard action payloads already include component metadata and fields—no breaking changes.
- Documentation should clarify that handlers may receive more frequent actions and should be idempotent.
- Handlers can inspect the trigger header to implement event-specific logic (e.g., lighter validation on `keyup`, full validation on `blur`).

## Additional Considerations

### Accessibility
- Trigger-driven updates should announce changes to screen readers using ARIA live regions.
- When a trigger fires, check if the updated component has `aria-live` or add `role="status"` to the target container.
- Document best practices for accessible trigger usage (e.g., debounce search to reduce announcement frequency).

### WebSocket Integration
- Triggers work over both HTTP and WebSocket transports automatically.
- When WebSocket is active, trigger requests use the WebSocket connection.
- The `X-Liveflux-Trigger` header is sent as part of the WebSocket message payload.
- Triggers are independent of `$wire` API—no `initWire` required.

### Memory Management
- Value cache is scoped per element using WeakMap to prevent memory leaks.
- Timers are cleared when elements are removed from DOM (via MutationObserver or explicit cleanup in `applyTargets`).
- Listener registry uses WeakMap to allow garbage collection of removed elements.

## Testing Strategy
- **Unit**: new Jasmine specs verifying debounce/throttle behavior, `changed` filter, `queue` strategies, and cleanup on DOM replacement.
- **Integration**: Go-side tests ensuring handler receives expected payloads when triggers fire.
- **Examples**: add `examples/triggers/` demonstrating live search, select-driven filters, blur validation.

## Documentation Plan
- Update `docs/components.md` and `docs/handler_and_transport.md` with trigger usage examples.
- Add a quick-start recipe in `docs/getting_started.md` for search-on-typing.
- Highlight interplay with `data-flux-select` and targeted fragment updates.

## Rollout Steps

### Phase 1: Core Functionality
1. Implement parser & listener registry in new module `js/liveflux_triggers.js`.
2. Add support for:
   - Event parsing and default inference
   - `delay` modifier with debounce
   - `changed` filter with value caching
   - `once` filter
   - `queue:replace` strategy (default)
3. Wire bootstrap + re-init hooks in `liveflux_bootstrap.js`.
4. Add `X-Liveflux-Trigger` header to requests.
5. Write unit tests (`js/tests/triggers.spec.js`).
6. Create example: `examples/triggers/search.go` (live search demo).
7. Document in `docs/triggers.md` with migration guide.

### Phase 2: Advanced Modifiers
1. Add `throttle` modifier.
2. Add `from:<selector>` and `not:<selector>` filters.
3. Add `queue:all` strategy.
4. Create example: `examples/triggers/filters.go` (select-driven filters).

### Phase 3: Shortcuts & Polish
1. Implement shortcut attributes (`data-flux-change`, `data-flux-keyup`, `data-flux-blur`).
2. Add `ClientOptions` for global defaults.
3. Create example: `examples/triggers/validation.go` (blur validation).
4. Beta test on internal components before public release.

## Resolved Design Decisions

### Global Configuration
Add `ClientOptions` support for:
- `defaultTriggerDelay`: Global minimum delay for all triggers (default `0ms`). Useful for rate limiting.
- `enableTriggers`: Boolean flag to disable trigger system for debugging (default `true`).

### Request Indicators
When multiple triggers exist on the same element, use the element itself for indicator placement. The indicator shows during any pending request from that element.

### Event Ordering & Conflicts
When multiple trigger definitions exist (comma-separated), they operate independently:
```html
<input data-flux-trigger="keyup delay:300ms, blur" data-flux-action="search" />
```
- `keyup` starts a 300ms timer; if `blur` fires before timer expires, both requests may fire (respects queue strategy).
- With `queue:replace`, the `blur` request cancels the pending `keyup` request.

### Native Event Conflicts
When both trigger and native behavior exist:
```html
<form data-flux-action="save" data-flux-trigger="submit">
  <button type="submit">Save</button>
</form>
```
- The trigger intercepts the native submit event using `preventDefault()`.
- Only the trigger-driven action fires; native form submission is suppressed.
- For progressive enhancement, omit `data-flux-trigger` to preserve native fallback.

### Error Handling
- If `data-flux-trigger` references a non-existent action, log a console warning and skip the trigger.
- If trigger fires but component metadata is missing, fall back to standard error handling (same as click actions).
- Malformed trigger syntax (e.g., `delay:invalid`) logs a warning and ignores the modifier.

## Alternatives Considered
- **Manual JS helpers**: flexible but contradicts Liveflux’s HTML-first philosophy.
- **Server opt-in**: requiring explicit component support adds friction and repeats logic across components.
- **Rely on WebSocket wire (`$wire.on`)**: works but requires `initWire` and custom JS, not declarative.
- **JSON-based trigger config**: more powerful but less readable than attribute-based syntax.

## Conclusion
Adding `data-flux-trigger` brings Liveflux closer to htmx ergonomics while preserving existing architecture. The change is client-side only, incremental, and unlocks a range of interactions (search-as-you-type, instant filters, autosave) with minimal effort from component authors.
