# Liveflux (Go) vs. Phoenix LiveView (Elixir) — Comparison

| Topic | Liveflux (Go) | Phoenix LiveView (Elixir) |
|---|---|---|
| Language / Stack | Go library | Phoenix/Elixir framework feature |
| Endpoint | POST/GET `/liveflux` | Router `live` routes; initial HTTP then WebSocket |
| Transport | Plain forms via built-in client (form-encoded) or optional WebSocket transport via `NewHandlerWS`/`NewWebSocketHandler` | WebSocket with JSON diff patches (after initial render) |
| Rendering | Server returns full HTML (`hb.TagInterface.ToHTML()`) | Server renders HEEx; client patches DOM via diffs |
| State | Component instance persisted via `Store` (default in-memory) | Server process holds `socket.assigns` (state on server) |
| Templating | `hb` (builder) by default; any HTML works | HEEx (HTML + Elixir), function components, slots |
| Client directives | Minimal (`data-flux-action`, placeholders) | Rich `phx-*` events (`phx-click`, `phx-submit`, `phx-change`, etc.) |
| Redirects | Custom redirect headers + HTML fallback | Live navigation: `push_patch`, `push_redirect`, `redirect` |
| SSR | Inherent (server-rendered each request) | Initial server render; then LiveView upgrades over WS |
| Partial updates | Template fragment targets (`data-flux-target`) with optional document-scoped selectors; falls back to full swap if selectors fail | DOM patches via diff protocol; `phx-update` modes |
| Two-way binding | Not built-in (manual via `Handle`) | Form syncing via `phx-change`/`phx-submit`, `phx-debounce`/`phx-throttle` |
| File uploads | Not built-in | Built-in Live Uploads with chunking/validation |
| CSRF | Add via normal forms/headers | Phoenix CSRF/auth tokens and signed sessions |
| Ecosystem | Lightweight, bring-your-own | Mature Phoenix ecosystem, telemetry, PubSub |

This document compares our Go package `liveflux` with Phoenix LiveView (Elixir), highlighting concepts, APIs, and trade-offs.

## Summary
- Liveflux is a minimal, server-driven component system with optional WebSocket support.
- Transport is simple form POST/GET via the included client script; rendering is HTML (via `hb`).
- Phoenix LiveView provides real-time, stateful UI over WebSockets with HEEx templates, rich client events, uploads, and first-class router/view integration.

## Architecture
- __Our pkg (`liveflux`)__
  - Endpoint: `POST/GET /liveflux` (`handler.go`).
  - Transport: built-in lightweight client (embedded via `script.go`), standard form-encoded POST/GET, or optional WebSocket transport via `NewHandlerWS`/`NewWebSocketHandler`.
  - Rendering: server returns full component HTML using `hb.TagInterface.ToHTML()`.
  - State: `Store` interface with default in-memory `MemoryStore` (`state.go`).
  - Registration: type-based registry, aliases via `Register`/`RegisterByAlias` (`registry.go`).
- __Phoenix LiveView__
  - Tight integration with Phoenix router, controllers, templates, and plugs.
  - Initial HTTP render (static or connected), then upgrade to a WebSocket for live diffs.
  - Rendering: server re-renders HEEx; client applies DOM patches based on diffs.
  - State: maintained on the server in the LiveView process (`socket.assigns`), with lifecycle callbacks.

## Component Model & Lifecycle
- __Our pkg__ (`component.go`)
  - Interface: `ComponentInterface` with `GetAlias/SetAlias`, `GetID/SetID`, `Mount(ctx, params) error`, `Handle(ctx, action string, data url.Values) error`, `Render(ctx) hb.TagInterface`.
  - Lifecycle: mount (first POST without `id`) -> state stored -> handle actions -> render.
  - Redirects: components can call `Base.Redirect(url, delaySeconds...)`. Handler emits custom redirect headers and a fallback HTML page (`handler.go`).
- __Phoenix LiveView__
  - Module-based LiveViews with callbacks: `mount/3`, `handle_params/3`, `handle_event/3`, `render/1`.
  - Rich lifecycle and state management via assigns, temporary assigns, and presence/streams.
  - Live Components (stateful and stateless) for composition.

## Client API & Templating
- __Our pkg__
  - Templating via `github.com/dracory/hb` (builder producing HTML). No HEEx-equivalent.
  - Mount via helper: `liveflux.PlaceholderByAlias(alias, params)` producing `<div data-flux-mount="1" data-flux-component="...">` consumed by the built-in client.
  - Actions: elements annotated with `data-flux-action`, posts include `component`, `id`, `action`.
- __Phoenix LiveView__
  - HEEx templating with assign interpolation and function components.
  - Rich client events and attributes: `phx-click`, `phx-submit`, `phx-change`, `phx-keydown`, `phx-blur`, `phx-focus`, `phx-target`, `phx-value-*`.
  - JS commands (`Phoenix.LiveView.JS`) for client-side interactions and transitions.

## State & Data Binding
- __Our pkg__
  - State is Go struct fields on the component instance; persisted in `Store`.
  - No automatic two-way binding. Inputs are serialized on action and passed as `url.Values` to `Handle()`.
- __Phoenix LiveView__
  - State lives in `socket.assigns`. Forms sync via `phx-change` (validate as you type) and `phx-submit`.
  - Debounce/throttle for inputs via `phx-debounce`/`phx-throttle` modifiers.
  - Temporary assigns to reduce payload size between renders.

## Transport & Protocol
- __Our pkg__ (`handler.go`)
  - Form fields: `component`, `id`, `action`.
  - Mount: POST/GET `component=alias` (+params) -> returns HTML.
  - Action: POST/GET `component, id, action` (+fields) -> returns HTML or redirect headers.
- __Phoenix LiveView__
  - Initial HTTP request returns rendered HTML + connection data; subsequent communication over WebSocket with a diff protocol.
  - Client applies minimal DOM patches; effects include push events, JS commands, navigation.

## SSR & Partial Updates
- __Our pkg__
  - SSR inherent (server-rendered each request).
  - Targeted fragment responses update only matching selectors; omit component metadata to patch document-scoped regions shared across components.
  - Automatic fallback to full component render when selectors fail or no fragments are returned.
- __Phoenix LiveView__
  - Server renders HEEx; client applies DOM patches via diffs for granular updates.
  - `phx-update` controls node replacement/append/prepend/ignore semantics.

## Features Comparison
- __Implemented in our pkg__
  - Server-driven components with `Mount/Handle/Render`.
  - Targeted fragment updates with component-scoped and document-scoped selectors.
  - Type-safe registry and aliasing helpers (`functions.go`: `DefaultAliasFromType`, `NewID`).
  - In-memory `Store` with pluggable interface (`state.go`).
  - Basic redirects with delay headers.
  - Minimal client: embedded JS (mount placeholders, action clicks, form submit, script re-execution).
  - Optional WebSocket transport with `WebSocketHandler`, including origin allow-listing, CSRF checks, TLS enforcement, rate limiting, and per-message validation (`websocket.go`).
- __Not (yet) implemented vs. Phoenix LiveView__
  - WebSocket transport with diff protocol and granular DOM patching.
  - Built-in form/state binding with debounce/throttle.
  - Live uploads with chunking and validations.
  - Live navigation (`push_patch`, `push_redirect`) and URL param syncing.
  - Streams and presence utilities for large lists and real-time feeds.
  - Built-in CSRF/session integration (can be added manually via forms/headers).

## Developer Experience
- __Our pkg__
  - Go-first types, explicit `Handle()` for actions, rendering with `hb`.
  - Very small surface area; easy to reason about and test.
  - You choose transport (included client or standard forms/fetch) and storage (swap `Store`).
- __Phoenix LiveView__
  - Convention-rich DX with router integration, HEEx, `phx-*` events, Live Components, and JS commands.
  - Strong docs and ecosystem; integrates with Phoenix generators, telemetry, and PubSub.

## Security Notes
- __Our pkg__
  - Add CSRF tokens if needed; `fetch`/standard forms accept typical tokens/headers.
  - Validate/authorize in `Handle()`; avoid sensitive data in client-visible fields.
- __Phoenix LiveView__
  - CSRF protection via Phoenix authenticity token; signed session data used to connect LiveViews.
  - Authorization/validation handled in `mount/3`, `handle_event/3`, or plugs.

## Performance Considerations
- __Our pkg__
  - Full component re-render and outerHTML swap per action. Simple and predictable; may be heavier than DOM-diffing for large trees.
  - Stateless HTTP with store lookup. For multi-instance deployments, provide a shared/session-backed `Store`.
- __Phoenix LiveView__
  - WebSocket diffs minimize payloads and layout thrashing; very low-latency interactions.
  - Temporary assigns and streams optimize large lists and reduce diff sizes.

## When to Choose Which
- __Choose Liveflux__
  - You’re in Go, prefer server-rendered UIs, and want minimal coupling and simple primitives.
  - You’re fine without two-way binding and advanced directives (or will implement them incrementally).
  - You want full control over transport and state storage.
- __Choose Phoenix LiveView__
  - You’re in Elixir/Phoenix and want stateful, real-time UI with minimal JavaScript and excellent developer ergonomics.

## Mapping Examples
- __Mounting__
  - Our: `liveflux.PlaceholderByAlias("counter")` -> client mounts via `/liveflux`.
  - Elixir: Router `live "/counter", MyAppWeb.CounterLive` and render via `live_render/3` or by navigating to the route.
- __Actions__
  - Our: button with `data-flux-action="inc"` -> `Handle(ctx, "inc", formValues)`.
  - Elixir: `<button phx-click="inc">` -> `handle_event("inc", params, socket)` increments state.
- __Redirect / Navigation__
  - Our: `c.Redirect("/next", 1)` -> custom redirect headers + fallback HTML.
  - Elixir: `push_patch(socket, to: "/next")` (same LiveView) or `push_redirect/redirect` for navigation.

## Gaps & Potential Roadmap
- Optional WebSocket channel with diffing for more granular updates.
- `wire:model`-like two-way binding helpers (client reads fields on input/change and submits diffs).
- Validation helpers with error bags and convenient rendering helpers in `hb`.
- Loading/disabled state helpers and progress indicators.
- File upload support (chunking and progress).
- Session-backed `Store` implementation and middleware example.
- Nested components with prop passing and event bubbling.

## References
- Our code: `component.go`, `handler.go`, `registry.go`, `state.go`, `placeholder.go`, `script.go`, `README.md`.
- Phoenix LiveView docs: https://hexdocs.pm/phoenix_live_view
