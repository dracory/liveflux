# Liveflux — Proposal for Improvements (Tech Lead Plan)

## Objectives
- Strengthen core primitives (simple, predictable server-driven components).
- Remove remaining legacy branding and align APIs/docs with “Liveflux”.
- Improve reliability (state, redirects, client), security guidance, and DX.
- Prepare for broader adoption with examples, tests, and versioned releases.

## Current State (summary)
- Components implement `ComponentInterface` with `Mount/Handle/Render` and alias/ID helpers (`component.go`).
- HTTP handler supports POST/GET, reads `component`, `id`, `action` (`handler.go`).
- Rendering uses `hb.TagInterface.ToHTML()`; output is plain HTML.
- Default in-memory `Store` (`state.go`); registry-based component lookup (`registry.go`).
- Minimal client shipped via `script.go`/`script.js` for mounting/actions.
- Docs largely up to date; README rebranded to “Liveflux” and HTMX dependency removed.

## Key Proposals

1) Client Script Enhancements (`script.go`/`js/*.js`)
- Add CSRF helper hooks: allow integrators to inject a token provider; auto-attach header/hidden field.
- Network resilience: simple retry/backoff on network failures; abort controller on rapid re-clicks.
- Progressive enhancement: run only when placeholders are present; avoid double-mount; re-exec inline scripts safely.
- Tiny event hooks: `lw:mounted`, `lw:beforeAction`, `lw:afterAction`, `lw:error` for integrators.

2) Security Guidance
- Document CSRF integration (examples for header + hidden input); mention rate limiting and auth checks in `Handle()`.
- Advise not to trust client values; validate/authorize on every action.
- Note XSS considerations when rendering user content with `hb` (escape, allowlists).