# Liveflux Data Attributes Reference

This page lists every `data-flux-*` attribute that the Liveflux runtime understands, grouped by feature area for quick lookup. Prefer the `data-flux-*` spellings shown here; legacy aliases without the `data-` prefix are still accepted for backwards compatibility unless noted otherwise.

## Component & Mounting

| Attribute | Purpose | Typical Placement |
| --- | --- | --- |
| `data-flux-component-kind="foo.bar"` | Identifies which component is mounted; together with `data-flux-component-id` this marks the root element. | Component root |
| `data-flux-component-id="abc123"` | Pairs with `data-flux-component-kind` so the runtime can look up the mounted instance. | Component root |
| `data-flux-id="…"` | Optional developer-defined identifier exposed on `$wire`. Useful for JS access. | Component root |
| `data-flux-mount="1"` | Marks placeholders the client should mount when bootstrapping. | Server-rendered placeholder containers |
| `data-flux-param-foo="bar"` | Provides initial mount parameters (become `params["foo"]` in `Mount`). | Roots/placeholders |
| `data-flux-dispatch-to="kind[:id]"` | Restricts client-dispatched events to a specific component kind or instance. | Elements calling `liveflux.dispatch*` |

## Trigger & Action Wiring

| Attribute | Purpose | Typical Placement |
| --- | --- | --- |
| `data-flux-action="save"` | Names the server action to invoke. | Buttons, links, form controls |
| `data-flux-trigger="input delay:300ms changed"` | Declaratively binds DOM events to `data-flux-action`; supports filters (`changed`, `once`, `from`, `not`) and modifiers (`delay`, `throttle`, `queue`). | Inputs, forms, custom controls |
| `data-flux-trigger-modifiers="…"` | Optional shorthand container for modifiers when using trigger shortcut attributes. | Same element as trigger |
| `data-flux-submit` | Marks a non-submit element that should behave like a submit button during posting. | Buttons/links |

## Form-less Data Collection & Indicators

| Attribute | Purpose | Typical Placement |
| --- | --- | --- |
| `data-flux-include="#selector, .other"` | Adds fields from outside the default scope into the payload. | Trigger elements |
| `data-flux-exclude=".sensitive"` | Removes fields from the payload after inclusion. | Trigger elements |
| `data-flux-indicator="#spinner, this"` | Elements that should show loading state (`flux-request` class) while a request runs. | Buttons, links, component roots |

## Targeted Updates & Partial Rendering

| Attribute | Purpose | Typical Placement |
| --- | --- | --- |
| `data-flux-select="#fragment"` | When a full render returns, only the selected fragment(s) replace the root. | Trigger elements |
| `data-flux-target="#cart-total"` | Inside response `<template>` elements, identifies the DOM node to update. | Server responses created by `TargetRenderer` |
| `data-flux-swap="replace | inner | beforebegin | afterbegin | beforeend | afterend"` | Controls how fragment content is merged with the target. | Same `<template>` as `data-flux-target` |
| `data-flux-component-kind` / `data-flux-component-id` (on `<template>`) | Metadata so the client validates the fragment belongs to the correct component instance. Set `TargetFragment.NoComponentMetadata=true` to omit for document-scoped swaps. | Same `<template>` |

## Transport & Runtime Configuration

| Attribute | Purpose | Typical Placement |
| --- | --- | --- |
| `data-flux-ws` | Signals that the component/element should communicate over WebSocket. | Component roots, triggers |
| `data-flux-ws-url="/liveflux/ws"` | Overrides the default WebSocket endpoint for the marked element/component. | Same element as `data-flux-ws` |

## Cross-Component Metadata Helpers

| Attribute | Purpose | Typical Placement |
| --- | --- | --- |
| `data-flux-component-type="foo.list"` | Lets elements outside a component root specify which component kind they target. | Buttons/links outside the root |

## Aliases & Compatibility Notes

- Legacy spellings without the `data-` prefix (for example `flux-action`) are still matched by the runtime but should be considered deprecated.
- Some older documentation and examples may show `data-flux-component="…"`; the preferred attributes are `data-flux-component-kind` and `data-flux-component-id`.
- For server-generated fragments, omit component metadata by enabling `TargetFragment.NoComponentMetadata` when you need document-scoped updates.
