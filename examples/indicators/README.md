# Liveflux Indicators Example

Demonstrates several request indicator patterns that toggle while the server processes actions.

## Run

```bash
# from the liveflux repo root
go run ./examples/indicators
```

Then open http://localhost:8084

## Components

1. **Button Demo (`IndicatorDemo`)**  
   Simulates a slow fetch and uses `data-flux-indicator="this, #demo-spinner"` so both the button and spinner animate. The spinner starts hidden (`display: none`) and is revealed automatically during the request.

2. **Form Demo (`IndicatorForm`)**  
   Submits a simple form and targets a spinner inside the submit button. It shows how form submissions can reuse the indicator helper while updating a status panel via `data-flux-select`.

3. **External Indicator (`ExternalIndicatorDemo`)**  
   Toggles the shared page-level indicator with `data-flux-indicator="#global-indicator"`, demonstrating that indicators do not need to live inside the component’s root.

Each component intentionally sleeps for ~1s so the indicator state is visible.

## Global indicator

The page hosts a reusable element with `id="global-indicator"`. Any component can reference it in `data-flux-indicator` to provide consistent UX for long-running actions (e.g., background tasks, cross-component coordination).

There's also a top-level button wired with `data-flux-target-kind` / `data-flux-target-id` to trigger the button demo's `fetch` action, showing how external controls can drive indicators on other components.
