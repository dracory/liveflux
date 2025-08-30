# Liveflux (Go) vs. StimulusReflex (Ruby) — Comparison

| Topic | Liveflux (Go) | StimulusReflex (Ruby on Rails) |
|---|---|---|
| Language / Stack | Go library | Rails ecosystem: Stimulus + Action Cable + Reflex |
| Endpoint | POST/GET `/liveflux` | Normal Rails routes; Reflex over WebSocket (Action Cable) |
| Transport | Plain forms via built-in client (form-encoded) | WebSocket messages; DOM morphs via morphdom |
| Rendering | Server returns full HTML (`hb.TagInterface.ToHTML()`) | Server re-renders ERB/partials; client morphs DOM |
| State | Component instance persisted via `Store` (in-memory default) | Request/session/Reflex instance vars; no long-lived component state |
| Templating | `hb` (builder) by default; any HTML works | Rails views/partials (ERB/HAML/SLIM), CableReady ops |
| Client directives | Minimal (`data-lw-action`, placeholders) | `data-reflex` attributes, Stimulus controllers, `stimulus_reflex` helpers |
| Redirects | Custom redirect headers + HTML fallback | Standard Rails redirects; Turbo-compatible setups often used |
| SSR | Inherent (server-rendered each request) | Inherent; Reflex augments with WS roundtrips |
| Partial updates | OuterHTML swap (no DOM diff) | morphdom-based granular DOM patching via HTML diffs |
| Two-way binding | Not built-in (manual via `Handle`) | No automatic two-way binding; Stimulus handles inputs |
| File uploads | Not built-in | Via standard Rails forms; not Reflex-specific |
| CSRF | Add via normal forms/headers | Rails authenticity token; Action Cable connection auth |
| Ecosystem | Lightweight, bring-your-own | Rails ecosystem; CableReady + StimulusReflex community |

This document compares our Go package `liveflux` with StimulusReflex, highlighting concepts, APIs, and trade-offs.

## Summary
- Liveflux: minimal server-driven components using HTTP form submissions and full-HTML returns.
- StimulusReflex: enhances Rails apps with real-time interactions using WebSockets; Reflex actions re-render partials and morph the DOM.

## Architecture
- __Our pkg (`liveflux`)__
  - Endpoint: `POST/GET /liveflux` (`handler.go`).
  - Transport: built-in lightweight client, form-encoded POST/GET.
  - Rendering: returns full component HTML via `hb.TagInterface.ToHTML()`.
  - State: `Store` interface with default `MemoryStore` (`state.go`).
  - Registration: component registry/aliases (`registry.go`).
- __StimulusReflex__
  - Built on Rails controllers/views with Stimulus controllers on the client.
  - Reflex actions are invoked over Action Cable; server re-renders HTML and sends patches to morphdom.
  - CableReady can perform targeted DOM operations beyond simple replacement.

## Component Model & Lifecycle
- __Our pkg__ (`component.go`)
  - Interface: `ComponentInterface` with `Mount`, `Handle(action, data)`, `Render`.
  - Lifecycle: mount (no `id`) -> store state -> handle actions -> re-render HTML -> outerHTML swap.
- __StimulusReflex__
  - Reflex classes (Ruby) define methods invoked by `data-reflex` triggers.
  - Lifecycle hooks: `before_reflex`, `around_reflex`, `after_reflex`, etc.
  - State typically lives in DB/session; Reflex instance vars are short-lived per interaction.

## Client API & Templating
- __Our pkg__
  - Mount via `liveflux.PlaceholderByAlias(alias, params)` producing `<div data-lw-mount>` consumed by our client.
  - Actions via `data-lw-action` sending `component`, `id`, `action`.
- __StimulusReflex__
  - Stimulus controllers attach `data-reflex="click->ExampleReflex#do_thing"` (and similar) to elements.
  - Templates are standard Rails ERB/partials; Reflex replaces/morphs designated elements.
  - CableReady operations can append/prepend/replace/update targets by selectors/ids.

## State & Data Binding
- __Our pkg__
  - State is Go struct fields persisted in a `Store`.
  - No automatic two-way binding; inputs serialized to `Handle()`.
- __StimulusReflex__
  - No built-in two-way binding; use Stimulus to read inputs and trigger Reflex.
  - Server determines new HTML from current params/session/DB and returns patches.

## Transport & Protocol
- __Our pkg__ (`handler.go`)
  - Mount: POST/GET `component=alias` (+params) -> HTML.
  - Action: POST/GET `component, id, action` -> HTML or redirect headers.
- __StimulusReflex__
  - Client sends Reflex payload over Action Cable identifying the target reflex, element state, and params.
  - Server returns updated HTML; client uses morphdom to patch the DOM.

## SSR & Partial Updates
- __Our pkg__
  - Full component subtree re-render and root `outerHTML` swap.
- __StimulusReflex__
  - Server renders partials; morphdom applies granular patches, minimizing layout thrash.
  - CableReady operations can target specific nodes for fine-grained updates.

## Features Comparison
- __Implemented in our pkg__
  - Server-driven components with explicit action handling.
  - Pluggable state store, minimal embedded JS client, redirect headers with fallback.
- __Not (yet) implemented vs. StimulusReflex__
  - WebSocket transport and morphdom-based granular patching.
  - CableReady-like client operation catalog (append/prepend/replace/update/etc.).
  - Built-in Stimulus bridge for declarative `data-*` triggers.

## Developer Experience
- __Our pkg__
  - Small, explicit API; bring your own patterns.
- __StimulusReflex__
  - Conventional Rails DX, Stimulus for small JS controllers, Reflex for server actions.
  - Rich lifecycle hooks; integrates well with Turbo/Action Cable.

## Security Notes
- __Our pkg__
  - Add CSRF tokens to forms/requests as needed.
- __StimulusReflex__
  - Inherits Rails CSRF protections and Action Cable authentication.

## Performance Considerations
- __Our pkg__
  - Full outerHTML swaps; simple but heavier than morph-based diffs on large DOMs.
- __StimulusReflex__
  - WebSocket transport with DOM morphing reduces payload and reflow.
  - CableReady targets specific nodes for minimal updates.

## When to Choose Which
- __Choose Liveflux__
  - Go stack, desire for explicit component state and simple HTTP transport.
- __Choose StimulusReflex__
  - Rails stack, prefer Stimulus controllers + server-side Reflex actions with minimal custom JS.

## Mapping Examples
- __Mounting__
  - Our: `liveflux.PlaceholderByAlias("counter")` -> client mounts via `/liveflux`.
  - Rails: normal ERB partial rendered in a view; Stimulus controller attaches behaviors.
- __Actions__
  - Our: button `data-lw-action="inc"` -> `Handle(ctx, "inc", formValues)`.
  - Rails: `<button data-reflex="click->CounterReflex#inc">` -> `CounterReflex#inc` updates and re-renders partial; DOM morph applied.
- __Redirects__
  - Our: `c.Redirect("/next", 1)` -> custom headers + fallback HTML.
  - Rails: standard `redirect_to` or Turbo navigation in combo setups.

## Gaps & Potential Roadmap
- Optional WS channel + DOM-diff/morph client.
- `data-*` directive layer bridging Stimulus-like triggers to Liveflux actions.
- Catalog of client-side operations akin to CableReady.
- Helpers for targeting DOM nodes and rendering partial fragments.

## References
- Our code: `component.go`, `handler.go`, `registry.go`, `state.go`, `placeholder.go`, `script.go`, `README.md`.
- StimulusReflex docs: https://docs.stimulusreflex.com
