# Liveflux Indicators Example

This example demonstrates several request indicator patterns and **where** indicators can live
relative to the component that triggers them (inside the component root vs elsewhere on the page).

## Run

```bash
# from the liveflux repo root
go run ./examples/indicators
```

Then open http://localhost:8084

## Components

There are three main demos rendered as Bootstrap cards:

1. **Fetch Data Demo (`fetchDataComponent`)**  
   *Where things live: both the trigger button and the spinner are inside the card.*  
   This component simulates a slow `fetch` action and uses
   `data-flux-indicator="this, .demo-spinner"` so **both** the button and the small spinner
   next to it animate. The spinner starts hidden (`display: none`) and is revealed automatically
   during the request.

2. **Form Demo (`IndicatorForm`)**  
   *Where things live: the text input, submit button, and its spinner are all inside the card.*  
   A simple form that posts a `name` field. The submit button uses
   `data-flux-indicator="this, #form-spinner"` so the button and its inline spinner animate while
   the form is being processed. The message panel above the form is updated via
   `data-flux-select="#form-status"`.

3. **External Indicator Demo (`ExternalIndicatorDemo`)**  
   *Where things live: the card contains only the trigger and status text, but it drives a
   **page-level** indicator rendered above all cards.*  
   The component’s button uses `data-flux-indicator="#global-indicator"` to toggle a shared banner
   at the top of the page. This demonstrates that indicators do **not** need to live inside the
   component’s root; any element on the page can be targeted.

Each component intentionally sleeps for around one second so the indicator state is easy to see.

## Global indicator and external triggers

The page hosts a reusable element with `id="global-indicator"` above the cards. Any component or
button can reference it in `data-flux-indicator` to provide a consistent, page-level UX for
long‑running actions (for example, background tasks or cross‑component coordination).

The main page also wires **external buttons** (outside any component card) using
`data-flux-target-kind` / `data-flux-target-id` to invoke the fetch action on the
`fetchDataComponent` while toggling either the global indicator or a local spinner next to the
button. This shows how **controls and indicators can live outside the card**, while still driving
actions and indicators for a specific component.
