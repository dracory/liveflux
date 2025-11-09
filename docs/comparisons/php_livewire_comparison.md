# Liveflux (Go) vs. Laravel (PHP) Livewire — Comparison

| Topic | Liveflux (Go) | Laravel Livewire (PHP) |
|---|---|---|
| Language / Stack | Go library | Laravel/PHP framework feature |
| Endpoint | POST/GET `/liveflux` | Laravel-defined route(s) |
| Transport | Plain forms via built-in client (form-encoded) or optional WebSocket transport via `NewHandlerWS`/`NewWebSocketHandler` | JSON payloads with diffs + DOM morph |
| Rendering | Server returns full HTML (`hb.TagInterface.ToHTML()`) | Blade views re-rendered; client morphs DOM |
| State | Component instance persisted via `Store` (default in-memory) | Public properties serialized (session-like) |
| Templating | `hb` (builder) by default; any HTML works | Blade templates |
| Client directives | Minimal (`data-flux-action`, placeholders) | Rich (`wire:*`, Alpine integration) |
| Redirects | Custom redirect headers + HTML fallback | Framework redirects |
| SSR | Inherent (server-rendered each request) | Server-rendered Blade + client morph |
| Partial updates | Template fragment targets (`data-flux-target`) with optional document-scoped selectors; falls back to full swap if selectors fail | DOM diff/morph for granular updates |
| Two-way binding | Not built-in (manual via `Handle`) | Yes (`wire:model` + modifiers) |
| File uploads | Not built-in | Built-in helpers |
| CSRF | Add via normal forms/headers | Laravel middleware |
| Ecosystem | Lightweight, bring-your-own | Mature, batteries included |

This document compares our Go package `liveflux` with Laravel Livewire (PHP), highlighting concepts, APIs, and trade-offs.

## Summary
- Liveflux is a minimal, server-driven component system with optional WebSocket support.
- Transport is simple form POST/GET via the included client script; rendering is HTML (via `hb`).
- Laravel Livewire is a mature framework integrated with Laravel + Blade, offering richer features (two-way binding, validation, uploads, nested layouts, Alpine integration, etc.).

## Architecture
- __Our pkg (`liveflux`)__
  - Endpoint: `POST/GET /liveflux` (`handler.go`).
  - Transport: built-in lightweight client (embedded via `script.go`), standard form-encoded POST/GET, or optional WebSocket transport via `NewHandlerWS`/`NewWebSocketHandler`.
  - Rendering: server returns full component HTML using `hb.TagInterface.ToHTML()`.
  - State: `Store` interface with default in-memory `MemoryStore` (`state.go`).
  - Registration: type-based registry, kinds via `Register`/`RegisterByKind` (`registry.go`).
- __Laravel Livewire__
  - Tight integration with Laravel (routing, middleware, CSRF, auth).
  - Transport: AJAX payloads with JSON diffs + DOM morphing.
  - Rendering: server re-renders Blade views; client morphs DOM.
  - State: component public properties serialized across requests; persistent via session-like mechanisms; lifecycle hooks are rich.

## Component Model & Lifecycle
- __Our pkg__ (`component.go`)
  - Interface: `ComponentInterface` with `GetKind/SetKind`, `GetID/SetID`, `Mount(ctx, params) error`, `Handle(ctx, action string, data url.Values) error`, `Render(ctx) hb.TagInterface`.
  - Lifecycle: mount (first POST without `id`) -> state stored -> handle actions -> render.
  - Redirects: components can call `Base.Redirect(url, delaySeconds...)`. Handler emits custom redirect headers and a fallback HTML page (`handler.go`).
- __Laravel Livewire__
  - Class-based components with `mount()`, `render()`, many lifecycle hooks (`boot`, `updating*`, `updated*`, etc.).
  - Strong conventions around public properties/events and validation.

## Client API & Templating
- __Our pkg__
  - Templating via `github.com/dracory/hb` (builder producing HTML). No Blade-equivalent.
  - Mount via our helper: `liveflux.PlaceholderByKind(kind, params)` which produces `<div data-flux-mount="1" data-flux-component-kind="...">` consumed by the built-in client.
  - Actions: buttons/submitters annotated with `data-flux-action`, posts include `component`, `id`, `action`.
- __Laravel Livewire__
  - Blade with `@livewire('component')` and directives.
  - Rich client directives: `wire:click`, `wire:submit`, `wire:model`, `wire:loading`, `wire:poll`, etc.
  - Deep Alpine.js interop and `$wire` bridge.

## State & Data Binding
- __Our pkg__
  - State is Go struct fields on the component instance; persisted in `Store`.
  - No automatic two-way binding. Inputs are serialized on action and passed as `url.Values` to `Handle()`.
- __Laravel Livewire__
  - Two-way binding via `wire:model` (with modifiers), syncing on specific triggers.
  - Public properties automatically serialized; nested arrays/objects supported.

## Transport & Protocol
- __Our pkg__ (`handler.go`)
  - Form fields: `component`, `id`, `action`.
  - Mount: POST/GET `component=kind` (+params) -> returns HTML.
  - Action: POST/GET `component, id, action` (+fields) -> returns HTML or redirect headers.
- __Laravel Livewire__
  - JSON payloads with component fingerprint, checksum, diffed properties, and effects; response includes DOM changes & effects.

## SSR & Partial Updates
- __Our pkg__
  - Targeted fragment responses update only matching selectors; omit component metadata when patching shared DOM outside a component root.
  - Automatic fallback to full component render when selectors fail or no fragments are returned.
- __Laravel Livewire__
  - Server renders Blade, client morphs DOM (partial updates), tracks effects; better incremental update performance on large trees.

## Features Comparison
- __Implemented in our pkg__
  - Server-driven components with `Mount/Handle/Render`.
  - Targeted fragment updates with component-scoped and document-scoped selectors.
  - Type-safe registry and kind helpers (`functions.go`: `DefaultKindFromType`, `NewID`).
  - In-memory `Store` with pluggable interface (`state.go`).
  - Basic redirects with delay headers.
  - Minimal client: embedded JS (mount placeholders, action clicks, form submit, script re-execution).
  - Optional WebSocket transport with `WebSocketHandler`, including origin allow-listing, CSRF checks, TLS enforcement, rate limiting, and per-message validation (`websocket.go`).
- __Not (yet) implemented vs. Laravel Livewire__
  - Two-way binding (`wire:model`), debouncing/throttling modifiers.
  - Built-in validation helpers integrated with form state.
  - File uploads, temporary file handling.
  - Loading/disabled states, progress indicators (`wire:loading`).
  - Polling, lazy/defer updates, entanglement with Alpine.
  - Nested component coordination (child props/events) beyond simple independent mounts.
  - DOM-diffing/morphing for granular updates.
  - Built-in CSRF integration (can be added via normal form mechanisms).

## Developer Experience
- __Our pkg__
  - Go-first types, explicit `Handle()` for actions, rendering with `hb`.
  - Very small surface area; easy to reason about and test.
  - You choose transport (included client or standard forms/fetch) and storage (swap `Store`).
- __Laravel Livewire__
  - Convention-rich DX with Blade and many directives; batteries included.
  - Strong community, docs, and ecosystem; deep Laravel integration.

## Security Notes
- __Our pkg__
  - Add CSRF tokens if needed; `fetch`/standard forms accept typical tokens/headers.
  - Validate/authorize in `Handle()`; avoid sensitive data in client-visible fields.
- __Laravel Livewire__
  - Inherits Laravel’s CSRF/auth middleware; validation helpers common.

## Performance Considerations
- __Our pkg__
  - Full component re-render and outerHTML swap per action. Simple and predictable; may be heavier than DOM-diffing for large trees.
  - Stateless HTTP with store lookup. For multi-instance deployments, provide a shared/session-backed `Store`.
- __Laravel Livewire__
  - Payload diffing and DOM morphing reduce transferred bytes & layout thrashing.

## When to Choose Which
- __Choose Liveflux__
  - You’re in Go, prefer server-rendered UIs, and want minimal coupling and simple primitives.
  - You’re fine without two-way binding and advanced directives (or will implement them incrementally).
  - You want full control over transport and state storage.
- __Choose Laravel Livewire__
  - You’re in Laravel/PHP with Blade and want a very productive, directive-rich system with mature features.

## Mapping Examples
- __Mounting__
  - Our: `liveflux.PlaceholderByKind("counter")` -> client mounts via `/liveflux`.
  - PHP: `@livewire('counter')` in Blade.
- __Actions__
  - Our: button with `data-flux-action="inc"` -> `Handle(ctx, "inc", formValues)`.
  - PHP: `<button wire:click="inc">` -> `increment()` method.
- __Redirect__
  - Our: `c.Redirect("/next", 1)` -> custom redirect headers + fallback HTML.
  - PHP: `return redirect('/next')` inside action.

## Gaps & Potential Roadmap
- Add `wire:model`-like two-way binding (client reads fields on input/change and submits diffs).
- Validation helpers with error bags and convenient rendering helpers in `hb`.
- Loading state helpers (attrs/classes while pending), disabled states.
- File upload support.
- Optional DOM-diffing/morphing client to reduce outerHTML swaps.
- Session-backed `Store` implementation and middleware example.
- Nested components with prop passing and event bubbling.

## References
- Our code: `component.go`, `handler.go`, `registry.go`, `state.go`, `placeholder.go`, `script.go`, `README.md`.
- Laravel Livewire docs: Components, wire:model, lifecycle hooks, reference.
