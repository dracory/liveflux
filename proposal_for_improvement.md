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

1) Branding & Redirect Headers
- Replace response headers `X-Livewire-Redirect` and `X-Livewire-Redirect-After` with `X-Liveflux-Redirect` and `X-Liveflux-Redirect-After` in `handler.go` and tests.
- Maintain backwards compatibility for one minor version (emit both; prefer Liveflux; add deprecation note in README/docs).
- Update `docs/` and examples accordingly.

2) Session-backed Store Implementation
- Provide `SessionStore` implementing `Store` (backed by cookie/session ID and server-side storage or pluggable interface).
- API: `NewSessionStore(getterSetter)` where app provides get/set by session key.
- Add docs with setup patterns (net/http middleware example) and trade-offs (expiry, size, scaling).

3) Client Script Enhancements (`script.go`/`js/*.js`)
- Add CSRF helper hooks: allow integrators to inject a token provider; auto-attach header/hidden field.
- Network resilience: simple retry/backoff on network failures; abort controller on rapid re-clicks.
- Progressive enhancement: run only when placeholders are present; avoid double-mount; re-exec inline scripts safely.
- Tiny event hooks: `lw:mounted`, `lw:beforeAction`, `lw:afterAction`, `lw:error` for integrators.

4) Security Guidance
- Document CSRF integration (examples for header + hidden input); mention rate limiting and auth checks in `Handle()`.
- Advise not to trust client values; validate/authorize on every action.
- Note XSS considerations when rendering user content with `hb` (escape, allowlists).

5) Testing & CI
- Increase unit tests around:
  - Redirect headers (new + legacy).
  - Store semantics (create/update/delete, concurrent access, TTL when applicable).
  - Client-side basic flows (mount, action, script re-execution) via lightweight DOM tests (go:embed HTML + headless runner or keep to integration tests via `net/http/httptest`).
- Ensure GitHub Actions matrix runs Go versions we support; add `-race` step.

6) API Ergonomics & DX
- Helper to simplify mounting params: `MountParams(params map[string]string) hb.TagInterface` or builder helpers on placeholders.
- Optional typed actions: sugar for mapping `action` to methods to reduce stringly-typed code.
- Generate default aliases consistently via `DefaultAliasFromType` and document the rule.

7) Documentation & Examples
- Add a plain HTML form + `fetch()` example (no `hb`) to README to prove framework-agnostic flow.
- Expand `docs/comparisons` with a quick migration guide mindset (how to think in Liveflux vs Livewire).
- Add a “Security” doc and a “State stores” doc (in-memory vs session vs custom).

8) Performance Notes
- Document costs of full outerHTML swaps and strategies:
  - Split into smaller components; avoid re-rendering big trees.
  - Use event delegation on client; limit script re-execution.
- Investigate optional DOM-diff/morph client as experimental (behind build flag).

9) Versioning & Deprecation Policy
- Semantic versioning: bump minor when adding features, major when removing legacy headers.
- Deprecation window for redirect header rename: 1 minor release.

## Phased Roadmap

- Phase 1 (Week 1)
  - Implement dual redirect headers (Liveflux + legacy).
  - Update tests (`handler_test.go`) and docs.
  - Add README plain HTML/fetch example.

- Phase 2 (Week 2)
  - Session-backed `Store` implementation + docs and example middleware.
  - Expand tests for store behavior.

- Phase 3 (Week 3)
  - Client script improvements: CSRF hooks, events, basic retries, no double mounts.
  - Add minimal client unit/integration tests where practical.

- Phase 4 (Week 4)
  - Security doc; state stores doc; polish comparisons.
  - CI: add `-race` job, coverage badge/report.

## Acceptance Criteria (per phase)
- Phase 1
  - New headers emitted and asserted in tests; README/docs updated with deprecation note.
  - No breaking changes (legacy headers still present).
- Phase 2
  - `SessionStore` passes tests; example app shows persistence across requests.
- Phase 3
  - Client emits lifecycle events; CSRF hook documented; retries observable in dev tools.
- Phase 4
  - Docs published; CI green across matrix; coverage unchanged or improved.

## Risks & Mitigations
- Header rename confusion — mitigate with dual emission + clear deprecation notes and changelog.
- Session store complexity — keep store interface narrow; provide sample middleware only.
- Client retries causing double actions — implement idempotency guidance and cancel-in-flight by `AbortController`.

## Open Questions
- Should we provide a reference session store using an external cache (Redis) for multi-instance deployments?
- Do we want an experimental DOM-diff client maintained in-tree or in a separate repo?
- What Go versions do we officially support in CI?
