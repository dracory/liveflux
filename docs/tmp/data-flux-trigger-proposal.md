# Proposal: Declarative Trigger Modifiers for Liveflux

## Summary
Introduce a `data-flux-trigger` attribute that lets authors declare client-side event bindings with optional filters (e.g. `changed`) and timing modifiers (e.g. `delay:300ms`, `throttle:1s`). The goal is to support patterns currently solved with custom JavaScript, such as firing component actions on `keyup` or `change` with debounce, while keeping the Liveflux runtime HTML-first and compatible with targeted fragment updates.

## Motivation
- **Parity with htmx**: teams moving from htmx expect `hx-trigger`-like ergonomics.
- **Lower JS footprint**: avoid hand-written listeners for common interactions (search-as-you-type, select-change refreshes, blur validation, etc.).
- **Progressive enhancement**: keep forms functional without JavaScript while enabling richer behavior when the client runtime is loaded.
- **Consistent API surface**: reuse Liveflux attribute naming (`data-flux-*`) and align with existing concepts like `data-flux-select`.

## Goals
1. Declaratively bind DOM events to Liveflux actions without extra scripts.
2. Support event filters (`changed`, `once`, `from:<selector>`) to reduce redundant requests.
3. Provide timing modifiers (`delay`, `throttle`, `queue`) to control frequency.
4. Maintain backwards compatibility—existing click and submit flows continue to work.
5. Integrate cleanly with targeted fragment updates and request indicators.

## Out of Scope (Initial Version)
- Custom user-defined modifiers beyond the built-in set.
- Complex gesture or multi-event sequences (e.g. `keyup changed, blur`).
- Streaming responses; focus on standard Liveflux postbacks.

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
- **Timing modifiers**: `delay:<duration>`, `throttle:<duration>`, `queue:<strategy>` (default `replace`).
- **Action resolution**: reuses `data-flux-action` on the same element; falls back to the nearest action definition in ancestors.

### Default Behavior
- When `data-flux-trigger` is present, Liveflux no longer requires a click/submit to fire the action.
- If both trigger and native submit exist, the first event to fire owns the request; subsequent triggers respect queuing rules.

### Form Integration
- For form controls, the runtime serializes fields using the existing `collectAllFields` helper so `data-flux-include` / `data-flux-exclude` continue to work.
- For standalone elements, runtime locates component metadata via `resolveComponentMetadata`, identical to click handling.

## Runtime Changes
1. **Parser**: add `liveflux.parseTriggers(el)` returning an array of `{ events, filters, modifiers }`.
2. **Listener registry**: maintain per-element listeners; ensure cleanup when DOM fragments are replaced (tie into `executeScripts` / `initWire`).
3. **Event handling**:
   - Evaluate filters (`changed` compares serialized value cache; `from`/`not` inspect `event.target`).
   - Apply timing modifiers using debounced / throttled wrappers (store per trigger state).
   - Respect `queue` strategies: `replace` (default), `all`, `first`, `last`.
4. **Request dispatch**: reuse `post` pipeline, including targeted templates, indicators, and error handling.
5. **Global bootstrap**: on `bootstrapInit`, scan for `[data-flux-trigger]` elements and register listeners. Re-run after any DOM update (`applyTargets`, `executeScripts`) to catch newly-inserted nodes.

## Server Considerations
- No protocol changes required; standard action payloads already include component metadata and fields.
- Documentation should clarify that handlers may receive more frequent actions and should be idempotent.

## Testing Strategy
- **Unit**: new Jasmine specs verifying debounce/throttle behavior, `changed` filter, `queue` strategies, and cleanup on DOM replacement.
- **Integration**: Go-side tests ensuring handler receives expected payloads when triggers fire.
- **Examples**: add `examples/triggers/` demonstrating live search, select-driven filters, blur validation.

## Documentation Plan
- Update `docs/components.md` and `docs/handler_and_transport.md` with trigger usage examples.
- Add a quick-start recipe in `docs/getting_started.md` for search-on-typing.
- Highlight interplay with `data-flux-select` and targeted fragment updates.

## Rollout Steps
1. Implement parser & listener registry in `js/liveflux_handlers.js` (or new module `liveflux_triggers.js`).
2. Wire bootstrap + re-init hooks.
3. Add debounce/throttle utilities (consider extracting small helper functions for reuse).
4. Write unit tests (`js/tests/triggers.spec.js`).
5. Document feature and update changelog.
6. Beta test on internal components before public release.

## Open Questions
- Should we support shorthand `data-flux-trigger="changed delay:500ms"` (implying default event based on element type)?
- Do we need server-provided defaults via `ClientOptions` (e.g., global minDelay)?
- How should multiple triggers on the same element interact with request indicators (use first matching, or all)?

## Alternatives Considered
- **Manual JS helpers**: flexible but contradicts Liveflux’s HTML-first philosophy.
- **Server opt-in**: requiring explicit component support adds friction and repeats logic across components.
- **Rely on WebSocket wire (`$wire.on`)**: works but requires `initWire` and custom JS, not declarative.

## Conclusion
Adding `data-flux-trigger` brings Liveflux closer to htmx ergonomics while preserving existing architecture. The change is client-side only, incremental, and unlocks a range of interactions (search-as-you-type, instant filters, autosave) with minimal effort from component authors.
