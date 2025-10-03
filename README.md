# Liveflux — Server-driven UI Components for Go

[![Tests Status](https://github.com/dracory/liveflux/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/dracory/liveflux/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/liveflux)](https://goreportcard.com/report/github.com/dracory/liveflux)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/liveflux)](https://pkg.go.dev/github.com/dracory/liveflux)

Liveflux is a server-driven component system for Go. It uses [`github.com/dracory/hb`](https://github.com/dracory/hb) for HTML generation but works with any server or frontend—the transport is plain HTML over HTTP or WebSocket.

## Highlights

- **Server-first rendering**: Components run on the server, returning HTML for any client.
- **Lightweight runtime**: Bundled JS handles mounts, actions, redirects, and optional WebSockets.
- **Composable state**: Per-component state persists via pluggable stores (`MemoryStore` by default).
- **Transport flexibility**: Use standard HTTP POST/GET or upgrade seamlessly to WebSockets.

## Quick start

- Install: `go get github.com/dracory/liveflux`
- Register a component and embed `liveflux.Base`
- Mount the handler at `/liveflux`
- Include `liveflux.Script()` and use `liveflux.PlaceholderByAlias()` or `liveflux.SSR()`

Full walkthroughs live in `docs/getting_started.md` and `docs/components.md`.

## Examples

### Counter (two instances side-by-side)

Run from repo root:

```bash
go run ./examples/counter
# or, with Task
task examples:counter:run
```

Screenshot:

![Counter Example](examples/counter/screenshot.png)

Source: `examples/counter/`

### Tree (nested list with modal add/edit)

Run from repo root:

```bash
go run ./examples/tree
# or, with Task
task examples:tree:run
```

Screenshots:

![Tree — Overview](examples/tree/screenshot.png)

![Tree — Edit node](examples/tree/screenshot_edit_node.png)

![Tree — Add child](examples/tree/screenshot_add_child.png)

Source: `examples/tree/`

### WebSocket Counter (two instances)

Run from repo root:

```bash
go run ./examples/websocket
# or, with Task
task examples:websocket:run
```

Screenshot:

![WebSocket Counter Example](examples/websocket/screenshot.png)

Source: `examples/websocket/`

## Documentation

Start with the focused guides under `docs/`:

- [Overview](docs/overview.md)
- [Getting started](docs/getting_started.md)
- [Components](docs/components.md)
- [Architecture](docs/architecture.md)
- [Handler & transport](docs/handler_and_transport.md)
- [State management](docs/state_management.md)
- [WebSocket integration](docs/websocket.md)

## Package API

- **Components**: `ComponentInterface`, `Base`, and registry helpers (`Register`, `RegisterByAlias`, `New`). See `docs/components.md` and `docs/architecture.md`.
- **Handlers**: HTTP + WebSocket entry points (`NewHandler`, `NewHandlerWS`, `NewWebSocketHandler`). See `docs/handler_and_transport.md` and `docs/websocket.md`.
- **SSR helpers**: `SSR`, `SSRHTML` for first render hydration. See `docs/ssr.md`.
- **Stores**: `Store` interface with default `MemoryStore`. See `docs/state_management.md` for custom implementations.

## Production notes

- Validate inputs in `Handle` and sanitize client data.
- Use session-backed or distributed stores in multi-instance deployments.
- Configure CSRF and allowed origins for HTTP/WebSocket endpoints.
- Decompose large views into smaller Liveflux components.

## License

This project is dual-licensed under the following terms:

- For non-commercial use, you may choose either the GNU Affero General Public License v3.0 (AGPLv3) _or_ a separate commercial license (see below). You can find a copy of the AGPLv3 at: https://www.gnu.org/licenses/agpl-3.0.txt

- For commercial use, a separate commercial license is required. Commercial licenses are available for various use cases. Please contact me via my [contact page](https://lesichkov.co.uk/contact) to obtain a commercial license.

## Similar projects

- Laravel Livewire (PHP): https://livewire.laravel.com [Comparison](docs/comparisons/php_livewire_comparison.md)
- Phoenix LiveView (Elixir): https://hexdocs.pm/phoenix_live_view [Comparison](docs/comparisons/phoenix_liveview_comparison.md)
- Hotwire Turbo (Rails): https://turbo.hotwired.dev [Comparison](docs/comparisons/hotwire_turbo_comparison.md)
- StimulusReflex (Ruby): https://docs.stimulusreflex.com [Comparison](docs/comparisons/stimulusreflex_comparison.md)
- Inertia.js (multi-framework): https://inertiajs.com
- Blazor (C#/.NET): https://dotnet.microsoft.com/apps/aspnet/web-apps/blazor [Comparison](docs/comparisons/blazor_comparison.md)
