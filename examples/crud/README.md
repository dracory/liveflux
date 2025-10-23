# Liveflux CRUD Example

A sophisticated example of using multiple liveflux components together on the same
page to create a CRUD (Create, Read, Update, Delete) application.

## Features
- Create User Modal Component
- Read Users List Component
- Update User Modal Component
- Delete User Confirmation Modal Component

## Run

```bash
# from the liveflux repo root
go run ./examples/crud
```

Then open http://localhost:8080

## Screenshot

![CRUD](./screenshot.png)

## What it does
- Renders a server-side CRUD using `liveflux.SSR`
- Mounts the JS runtime via `liveflux.Script()`
- Sends actions to `/liveflux` endpoint

## Notes
- The component uses `c.Root(content)` (provided by `liveflux.Base`) to wrap its markup with the standard Liveflux root and required hidden fields (`component`, `id`).
