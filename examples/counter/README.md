# Liveflux Counter Example

A minimal, runnable Liveflux example.

## Run

```bash
# from the liveflux repo root
go run ./examples/counter
```

Then open http://localhost:8080

## Screenshot

![Counter — Two instances](./screenshot.png)

## What it does
- Renders a server-side Counter using `liveflux.SSR`
- Mounts the JS runtime via `liveflux.Script()`
- Sends actions to `/liveflux` endpoint

## Notes
- The component uses `c.Root(content)` (provided by `liveflux.Base`) to wrap its markup with the standard Liveflux root and required hidden fields (`component`, `id`).
