# Liveflux (Go) vs. Hotwire Turbo (Rails) — Comparison

| Topic | Liveflux (Go) | Hotwire Turbo (Rails) |
|---|---|---|
| Language / Stack | Go library | Rails ecosystem (Turbo + Stimulus) |
| Endpoint | POST/GET `/liveflux` | Rails routes; normal controllers/views |
| Transport | Plain forms via built-in client (form-encoded) or optional WebSocket transport via `NewHandlerWS`/`NewWebSocketHandler` | HTML-over-the-wire; Turbo Frames/Streams; Action Cable for streams |
| Rendering | Server returns full HTML (`hb.TagInterface.ToHTML()`) | Server renders ERB/HTML; client updates frames/targets |
| State | Component instance persisted via `Store` (in-memory by default) | No persistent component state; request-driven; session for auth |
| Templating | `hb` (builder) by default; any HTML works | Rails views/partials (ERB/HAML/SLIM), `turbo-frame`, `turbo-stream` |
| Client directives | Minimal (`data-flux-action`, placeholders) | `data-turbo`, `turbo-frame`, `turbo-stream`, Stimulus controllers |
| Redirects | Custom redirect headers + HTML fallback | Standard Rails redirects; Turbo drive handles seamlessly |
| SSR | Inherent (server-rendered each request) | Inherent SSR; progressive enhancement by Turbo |
| Partial updates | OuterHTML swap (no DOM diff) | Targeted updates via Turbo Streams (append/prepend/replace/remove) |
| Two-way binding | Not built-in (manual via `Handle`) | No two-way binding; forms + Turbo Drive/Frames/Streams |
| File uploads | Not built-in | Standard Rails forms; Turbo-compatible |
| CSRF | Add via normal forms/headers | Rails authenticity token in forms/headers |
| Ecosystem | Lightweight, bring-your-own | Mature Rails ecosystem; Stimulus for JS behavior |

This document compares our Go package `liveflux` with Hotwire Turbo (Rails), focusing on concepts, APIs, and trade-offs.

## Summary
- Liveflux: minimal server-driven components with explicit `Mount/Handle/Render` cycle, simple HTTP transport, and optional WebSocket support.
- Hotwire Turbo: enhances traditional Rails apps with HTML-over-the-wire, Turbo Drive (navigation), Turbo Frames (partial page updates), and Turbo Streams (server-pushed updates).

## Architecture
- __Our pkg (`liveflux`)__
  - Endpoint: `POST/GET /liveflux` (`handler.go`).
  - Transport: built-in lightweight client, form-encoded POST/GET, or optional WebSocket transport via `NewHandlerWS`/`NewWebSocketHandler`.
  - Rendering: returns full component HTML via `hb.TagInterface.ToHTML()`.
  - State: `Store` interface with default `MemoryStore` (`state.go`).
  - Registration: component registry/aliases (`registry.go`).
- __Hotwire Turbo__
  - Works with standard Rails MVC: controllers render HTML; Turbo progressively enhances navigation and updates.
  - Turbo Drive: intercepts links/forms for fast navigation and partial reloads.
  - Turbo Frames: scoped partial updates; server responds with frame content to replace that frame.
  - Turbo Streams: server emits `<turbo-stream>` actions over HTTP or Action Cable for real-time updates.

## Component Model & Lifecycle
- __Our pkg__ (`component.go`)
  - Interface: `ComponentInterface` with `Mount`, `Handle(action, data)`, `Render`.
  - Lifecycle: mount (no `id`) -> store state -> handle actions -> re-render HTML -> outerHTML swap.
- __Hotwire Turbo__
  - Not a component runtime. Lifecycle follows standard Rails actions.
  - UI pieces are ERB partials. Turbo Frames scope refresh areas; Streams patch targets by `id` or `dom_id`.
  - Stimulus controllers add client behaviors where needed.

## Client API & Templating
- __Our pkg__
  - `liveflux.PlaceholderByAlias(alias, params)` renders `<div data-flux-mount>` consumed by the client.
  - Actions via `data-flux-action` (clicks/forms) posting `component`, `id`, `action`.
- __Hotwire Turbo__
  - Templates: Rails ERB with `turbo-frame id="..."` and partials.
  - Streams: server renders `<turbo-stream action="replace" target="...">` wrappers around partial HTML.
  - Attributes: `data-turbo`, `data-turbo-method`, `data-turbo-confirm`, etc.
  - Navigation: Turbo Drive replaces full/partial body content without full reload.

## State & Data Binding
- __Our pkg__
  - State is Go struct fields, persisted in `Store` across requests.
  - No built-in two-way binding; serialize inputs to `Handle()`.
- __Hotwire Turbo__
  - Stateless per request on the server (beyond Rails session). State re-derived from DB/session/params.
  - No automatic binding; forms post normally; validations re-render forms/partials which Turbo updates.

## Transport & Protocol
- __Our pkg__ (`handler.go`)
  - Mount: POST/GET `component=alias` (+params) -> HTML.
  - Action: POST/GET `component, id, action` -> HTML or redirect headers.
- __Hotwire Turbo__
  - Turbo Drive intercepts HTTP navigation and form submissions, expecting HTML response.
  - Turbo Frames: requests scoped to a frame; response HTML replaces that frame.
  - Turbo Streams: `<turbo-stream>` tags delivered via HTTP or Action Cable; actions `append`, `prepend`, `replace`, `remove`, `update`.

## SSR & Partial Updates
- __Our pkg__
  - Full component subtree re-render and root `outerHTML` swap.
- __Hotwire Turbo__
  - Server renders HTML; client applies targeted updates to frames/stream targets.
  - No JSON diff protocol; patches are HTML fragments indicated by stream actions.

## Features Comparison
- __Implemented in our pkg__
  - Server-driven components with explicit action handling.
  - Pluggable state store, minimal embedded JS client, redirect headers with fallback.
- __Not (yet) implemented vs. Hotwire Turbo__
  - First-class frame scoping for partial updates.
  - Stream actions with server-pushed updates (via SSE/WebSocket/Action Cable equivalent).
  - Turbo-like navigation layer for fast page transitions.
  - Built-in Rails-style helpers for DOM targeting (`dom_id`) and partial conventions.

## Developer Experience
- __Our pkg__
  - Small API surface; full control over transport and storage.
- __Hotwire Turbo__
  - Conventional Rails DX leveraging partials and helpers; Stimulus for sprinkles of JS.
  - Minimal custom JS; relies on HTML responses and declarative attributes.

## Security Notes
- __Our pkg__
  - Add CSRF tokens to forms/requests as needed.
- __Hotwire Turbo__
  - Inherits Rails CSRF protections (authenticity token); standard Rails auth/authorization patterns.

## Performance Considerations
- __Our pkg__
  - Full outerHTML swaps; simple but potentially heavier than targeted updates on large trees.
- __Hotwire Turbo__
  - HTML fragment updates avoid full reloads; streams enable real-time multi-client updates.
  - Caching partials and ETags/Conditional GET can further optimize.

## When to Choose Which
- __Choose Liveflux__
  - Go stack, server-rendered UIs, desire for explicit component state and simple primitives.
- __Choose Hotwire Turbo__
  - Rails stack, prefer HTML responses with minimal JS, want frames/streams for fast UX without a single-page app.

## Mapping Examples
- __Mounting__
  - Our: `liveflux.PlaceholderByAlias("counter")` -> client mounts via `/liveflux`.
  - Rails: `turbo-frame id="counter"` around content; controller renders the frame's partial.
- __Actions__
  - Our: button `data-flux-action="inc"` -> `Handle(ctx, "inc", formValues)`.
  - Rails: form submission to a controller action returns a `<turbo-stream action="replace" target="counter">` with updated partial.
- __Redirects__
  - Our: `c.Redirect("/next", 1)` -> custom headers + fallback HTML.
  - Rails: standard `redirect_to` works seamlessly with Turbo Drive.

## Gaps & Potential Roadmap
- Frame-like scoping helpers for Liveflux to limit update regions.
- Stream-like server-push with actions and HTML fragments.
- Navigation helper to emulate Turbo Drive behavior.
- View helpers to simplify partial rendering and DOM targeting.

## References
- Our code: `component.go`, `handler.go`, `registry.go`, `state.go`, `placeholder.go`, `script.go`, `README.md`.
- Hotwire Turbo docs: https://turbo.hotwired.dev
