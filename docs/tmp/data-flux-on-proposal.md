# Proposal: Inline Event Handlers via `data-flux-on`

## Summary
Add a `data-flux-on` attribute that lets authors attach lightweight client-side event handlers without leaving HTML. The goal is to cover the "run a little JS on this event" use cases that htmx solves with `hx-on`, while keeping Liveflux ergonomics consistent and safe. Handlers will execute in the context of the Liveflux namespace, exposing helper functions (dispatching events, reading metadata, calling `$wire`, etc.) without requiring globals.

## Motivation
- **Complement `data-flux-trigger`**: declarative triggers submit actions; inline handlers enable client-side side effects (analytics, UI tweaks) triggered by the same events.
- **Reduce custom scripts**: avoid scattering `<script>` tags or external bundles for simple behaviors (toggling classes, logging, emitting Liveflux events).
- **Improve composability**: allow component authors to ship self-contained markup that reacts to interactions without polluting global scope.
- **Parity with htmx `hx-on`**: makes migration/interop easier for teams familiar with the pattern.

## Goals
1. Provide a declarative way to run JS snippets when events fire on an element.
2. Keep handlers sandboxed—no direct `eval` of arbitrary global code; expose a curated API surface.
3. Support multiple event bindings per element and multiple statements per handler.
4. Integrate with Liveflux lifecycle (re-register handlers after DOM swaps, respect component teardown).
5. Ensure handlers can interact with Liveflux helpers (dispatch events, call actions, toggle indicators).

## Non-Goals
- Building a full scripting language—handlers remain short JS expressions/statements.
- Automatic minification or bundling of handler code.
- Server-driven execution of handlers.

## Authoring API
```html
<button
  data-flux-action="save"
  data-flux-trigger="click"
  data-flux-on='click: liveflux.events.dispatch("analytics:click", { id: this.dataset.id })'>
  Save
</button>
```

### Syntax Rules
- Attribute value is a semicolon-separated list of `event: expression` pairs.
- Events mirror DOM event names (`click`, `change`, `keyup`, `liveflux:rendered`, etc.).
- Expressions execute in a restricted scope with:
  - `this`: bound to the element.
  - `event`: the DOM event object.
  - `liveflux`: the global namespace.
  - Helper shortcuts: `dispatch`, `callAction`, `setIndicator`, etc.
- Multi-line handlers allowed via HTML entity encoding or `[[]]` block syntax (TBD).

### Example Use Cases
- Analytics hooks: `data-flux-on='click: liveflux.dispatch("track", {action: "save"})'`
- UI tweaks: `data-flux-on='mouseenter: this.classList.add("hover"); mouseleave: this.classList.remove("hover")'`
- Event bridging: `data-flux-on='change: $wire.call("refresh", {filter: event.target.value})'`

## Runtime Changes
1. **Parser**: `parseFluxOn(el)` returns `{ eventName, handlerFn }[]`. Parsing should support escaping colons and semicolons.
2. **Execution context**: create a sandboxed function wrapper that receives (`event`, `el`, `ctx`). `ctx` includes helpers:
   - `dispatch(name, data)` => `liveflux.dispatch(name, data)`
   - `$wire` => element’s wire proxy (if available)
   - `callAction(action, extraParams)` => run Liveflux action via `post`
   - `setIndicator(state)` => toggle request indicators
3. **Listener registry**: store references for cleanup on DOM replacement. Hook into existing `executeScripts` / `applyTargets` flows.
4. **Security**: avoid raw `eval`. Use `new Function` with explicit arguments and strip `return` statements? Document that handlers run with full JS privileges (like inline `onclick`).
5. **Error handling**: wrap execution in `try/catch`, log descriptive errors with element info.

## Interaction with Other Features
- **`data-flux-trigger`**: events can trigger both handlers and actions; order is handlers first, then trigger dispatch (configurable?).
- **Indicators**: provide helper to manually start/stop indicators for custom async work.
- **Wire**: expose `$wire` to call server actions or emit events.
- **Targeted updates**: ensure handlers re-register after component re-render.

## Documentation Plan
- Add section to `docs/components.md` describing `data-flux-on`, including scope and safety considerations.
- Provide recipes: analytics, UI tweaks, Liveflux event bridge.
- Compare/contrast with `data-flux-trigger` and when to use each.

## Testing Strategy
- **Unit tests** (`js/tests/on.spec.js`): parsing, execution context, multiple handlers, cleanup on DOM swap.
- **Integration**: ensure handlers survive targeted fragment updates and can call `$wire`.
- **Security**: tests for escaping `:` and `;`, invalid syntax handling.

## Rollout Steps
1. Implement parser and helper execution context in new `liveflux_on.js` module.
2. Register bootstrap hook to scan DOM for `[data-flux-on]` elements.
3. Integrate with re-init pipeline (post `applyTargets`, `executeScripts`).
4. Add unit tests and docs.
5. Update changelog and migration guide.

## Open Questions
- Should we allow referencing global functions (e.g., `window.analytics.track`)? Recommended via `liveflux.callGlobal('analytics.track', ...)` helper?
- How to handle async handlers—auto-wrap Promises to toggle indicators?
- Should handlers be able to cancel subsequent triggers by returning `false`?

## Alternatives Considered
- Encourage authors to use standard `addEventListener` scripts (more verbose, scatters logic).
- Rely solely on `data-flux-trigger` + server actions (cannot cover pure client logic).

## Conclusion
`data-flux-on` provides a balanced, HTML-first approach for inline event handling, complementing Liveflux’s server-driven model while keeping developers productive and codebases maintainable.
