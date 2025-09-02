# Liveflux (Go) vs. Blazor (C#/.NET) — Comparison

| Topic | Liveflux (Go) | Blazor (C#/.NET) |
|---|---|---|
| Language / Stack | Go library | .NET (ASP.NET Core) UI framework |
| Endpoint | POST/GET `/liveflux` | ASP.NET Core endpoints; Blazor Server (SignalR) or Blazor WebAssembly (HTTP) |
| Transport | Plain forms via built-in client (form-encoded) | Blazor Server: SignalR (WebSocket) diff patches; Blazor WASM: client-side runtime re-renders locally |
| Rendering | Server returns full HTML (`hb.TagInterface.ToHTML()`) | Razor Components render to DOM; diff-based renderer |
| State | Component instance persisted via `Store` (in-memory default) | Component state in .NET objects; cascading values/DI; persists in server circuit (Server) or in browser (WASM) |
| Templating | `hb` (builder) by default; any HTML works | Razor syntax (`.razor`), components, parameters, slots (`RenderFragment`) |
| Client directives | Minimal (`data-flux-action`, placeholders) | Event binding `@onclick`, `@onchange`, two-way `@bind`, lifecycle methods |
| Redirects | Custom redirect headers + HTML fallback | NavigationManager for client-side navigation; server redirects via ASP.NET |
| SSR | Inherent (server-rendered each request) | Optional: Blazor Server is stateful over SignalR; Blazor WASM is client-rendered; .NET 8+ supports SSR/streaming for Razor Components |
| Partial updates | OuterHTML swap (no DOM diff) | Diff/virtual DOM-like renderer applies minimal DOM patches |
| Two-way binding | Not built-in (manual via `Handle`) | Yes, `@bind` with format/culture/modifiers |
| File uploads | Not built-in | Built-in `<InputFile>` component and streaming APIs |
| CSRF | Add via normal forms/headers | ASP.NET Core antiforgery for forms; auth via Identity/AuthN/AuthZ |
| Ecosystem | Lightweight, bring-your-own | Extensive .NET ecosystem, tooling, components |

This document compares our Go package `liveflux` with Blazor, highlighting concepts, APIs, and trade-offs.

## Summary
- Liveflux: minimal server-driven components with simple HTTP transport and full-HTML responses.
- Blazor: a comprehensive component model in .NET, with Blazor Server (real-time over SignalR) and Blazor WebAssembly (runs .NET in the browser). Rich templating, binding, and tooling.

## Architecture
- __Our pkg (`liveflux`)__
  - Endpoint: `POST/GET /liveflux` (`handler.go`).
  - Transport: built-in lightweight client, form-encoded POST/GET.
  - Rendering: returns full component HTML via `hb.TagInterface.ToHTML()`.
  - State: `Store` interface with default `MemoryStore` (`state.go`).
  - Registration: type registry/aliases (`registry.go`).
- __Blazor__
  - Blazor Server: initial HTTP then a persistent SignalR circuit; server renders diffs and sends patches to the browser.
  - Blazor WebAssembly: .NET runtime downloads to the browser; all rendering and state are client-side.
  - Razor Components unify the model; can be hosted in ASP.NET Core with routing, layouts, DI, and JS interop.

## Component Model & Lifecycle
- __Our pkg__ (`component.go`)
  - Interface: `ComponentInterface` with `Mount(ctx, params)`, `Handle(ctx, action, data)`, `Render(ctx)`.
  - Lifecycle: mount -> persist state -> handle actions -> re-render -> outerHTML swap.
- __Blazor__
  - Components are `.razor` files/classes with parameters, `OnInitialized{Async}`, `OnParametersSet{Async}`, `OnAfterRender{Async}`.
  - Event handlers (`@onclick="Increment"`) mutate fields; `StateHasChanged()` schedules re-render.
  - Component composition via child components, parameters, cascading values, and slots (`RenderFragment`).

## Client API & Templating
- __Our pkg__
  - Templating via `github.com/dracory/hb` (HTML builder). Mount placeholders via `liveflux.PlaceholderByAlias(alias, params)` and `data-flux-action` for actions.
- __Blazor__
  - Razor templating combines HTML and C# inline; components encapsulate UI and logic.
  - Event binding with `@on*`, two-way binding with `@bind-Value`/`@bind`.
  - JS interop to call into JavaScript and receive callbacks.

## State & Data Binding
- __Our pkg__
  - State is Go struct fields per component instance; persisted in `Store` between requests.
  - No automatic two-way binding; inputs serialized to `Handle()`.
- __Blazor__
  - Strong two-way binding with validation via Data Annotations and EditForm/Input* components.
  - Dependency Injection and Cascading Parameters propagate state.
  - Blazor Server keeps state in memory per circuit; WASM keeps state in browser memory.

## Transport & Protocol
- __Our pkg__ (`handler.go`)
  - Mount: POST/GET `component=alias` (+params) -> HTML.
  - Action: POST/GET `component, id, action` -> HTML or redirect headers.
- __Blazor__
  - Server: SignalR connection carries UI diffs and events; HTTP for initial bootstrapping.
  - WASM: no server roundtrips for UI; HTTP for data APIs.

## SSR & Partial Updates
- __Our pkg__
  - Full component subtree re-render and root `outerHTML` swap.
- __Blazor__
  - Renderer computes minimal DOM patches and applies them efficiently.
  - .NET 8+ Razor Components can do SSR/streaming/hybrid rendering scenarios.

## Features Comparison
- __Implemented in our pkg__
  - Server-driven components with explicit action handling.
  - Pluggable state store, minimal embedded JS client, redirect headers with fallback.
- __Not (yet) implemented vs. Blazor__
  - Diff-based renderer with minimal DOM patching.
  - Rich two-way binding and validation components.
  - Persistent real-time circuit (SignalR) or full client-side runtime.
  - Built-in routing/layouts/component libraries.

## Developer Experience
- __Our pkg__
  - Small surface area; clear primitives; framework-agnostic.
- __Blazor__
  - First-class tooling (Visual Studio, Hot Reload), component libraries, strong typing and IDE support.

## Security Notes
- __Our pkg__
  - Add CSRF tokens to forms/requests as needed.
- __Blazor__
  - ASP.NET Core identity/auth, antiforgery for forms, authorization policies.

## Performance Considerations
- __Our pkg__
  - Full outerHTML swaps; simple but heavier than diff renderers on complex trees.
- __Blazor__
  - Efficient diffing; Blazor Server latency depends on network (SignalR roundtrips).
  - WASM downloads a runtime; after startup, interactions are local.

## When to Choose Which
- __Choose Liveflux__
  - Go stack, desire for simple server-controlled UI without heavy runtime.
- __Choose Blazor__
  - .NET stack, want a comprehensive component model with strong tooling and either real-time server rendering or full client-side execution.

## Mapping Examples
- __Mounting__
  - Our: `liveflux.PlaceholderByAlias("counter")` -> client mounts via `/liveflux`.
  - Blazor: Define `Counter.razor`; route via `@page "/counter"`; render as `<Counter />`.
- __Actions__
  - Our: button `data-flux-action="inc"` -> `Handle(ctx, "inc", formValues)`.
  - Blazor: `<button @onclick="Increment">` -> C# method mutates state; re-render occurs.
- __Redirects / Navigation__
  - Our: `c.Redirect("/next", 1)` -> custom headers + fallback HTML.
  - Blazor: `NavigationManager.NavigateTo("/next")` or server-side redirects in controllers.

## Gaps & Potential Roadmap
- Optional diffing renderer for granular DOM patches.
- `@bind`-like helpers for two-way form binding with validation hooks.
- Optional SignalR/WebSocket channel for real-time updates.
- Component/layout helpers and router integration samples.

## References
- Our code: `component.go`, `handler.go`, `registry.go`, `state.go`, `placeholder.go`, `script.go`, `README.md`.
- Blazor overview: https://dotnet.microsoft.com/apps/aspnet/web-apps/blazor
