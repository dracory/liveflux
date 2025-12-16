# Proposal: Cross-Component Action Targeting via `data-flux-target-*`

## Summary

Liveflux currently routes an action request to the **closest** component root element (an element with both `data-flux-component-kind` and `data-flux-component-id`). The runtime also supports `data-flux-target-kind` and `data-flux-target-id`, but only as a **fallback** when the trigger element is outside any component root.

This proposal introduces an explicit, ergonomic way to send an action to a different component instance **even when the trigger element is inside another component**.

## Motivation

A common pattern in server-driven UIs is a page composed of multiple components:

- A list/table component renders rows and action buttons.
- A modal component is rendered elsewhere on the page (or alongside the list) and is opened/closed via actions.

Today, if the modal open button is rendered inside the list component’s markup, clicking it will send:

- `liveflux_component_kind=<list kind>`
- `liveflux_component_id=<list id>`

…even if the button includes:

- `data-flux-target-kind=<modal kind>`
- `data-flux-target-id=<modal id>`

This forces one of the following workarounds:

- Dispatch an event from the list component and have the modal listen.
- Move the triggering element outside the list component root (often awkward for layout/templating).

Adding first-class cross-component action targeting reduces boilerplate and enables more natural component composition.

## Current Behavior (Implementation Notes)

The current runtime resolves metadata roughly as:

1. `btn.closest([data-flux-component-kind][data-flux-component-id])`
2. If not found, read `data-flux-target-kind` / `data-flux-target-id` from the trigger.

Because (1) succeeds for most triggers rendered inside components, (2) is rarely used.

## Proposed Behavior

### Baseline Proposal (Prefer explicit target)

If a trigger element has either:

- `data-flux-target-kind`
- `data-flux-target-id`

…then the runtime MUST treat that as an explicit override and route the request accordingly, regardless of whether the trigger is inside another component root.

Because the server expects both kind and id, the runtime MUST still send both values:

- If one of the target attributes is missing and the trigger is inside a component root, the runtime MAY fill the missing value from the closest component root.
- If the missing value cannot be resolved, the runtime MUST NOT fall back to the closest root silently.

### Attributes

No new attributes are required.

This proposal upgrades the semantics of existing attributes:

- `data-flux-target-kind`
- `data-flux-target-id`

From: “only useful outside component roots”

To: “explicit override for request routing”

### Example

```html
<!-- Inside providers list component root -->
<button
  type="button"
  class="btn btn-primary"
  data-flux-action="open"
  data-flux-target-kind="user_providers_modal_create"
  data-flux-target-id="qlLLpBVyhYbP"
  data-flux-indicator="this"
>
  New Provider
</button>
```

Expected outgoing payload:

- `liveflux_component_kind=user_providers_modal_create`
- `liveflux_component_id=qlLLpBVyhYbP`
- `liveflux_action=open`

## Backwards Compatibility

It may be breaking change. This is fine.

## Implementation Sketch (Client)

Update `resolveComponentMetadata(btn, rootSelector)` in `js/liveflux_util.js`:

- Check `data-flux-target-kind` and `data-flux-target-id` first.
- If both exist, return them.
- Otherwise, use closest root logic.

## Testing

Add/adjust JS tests to cover:

- Trigger inside a component root without `data-flux-target-*` routes to closest root.
- Trigger inside a component root with `data-flux-target-*` routes to target.
- Trigger outside any component root with `data-flux-target-*` still works.

## Documentation Updates

Update `docs/data_attributes_reference.md` and any relevant guides to reflect:

- `data-flux-target-kind` / `data-flux-target-id` can be used as an explicit routing override.

## Open Questions

- Should `data-flux-target-*` be respected for form submits as well (submitter button case)? Likely yes for consistency. Yes always override
