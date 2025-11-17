package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

func main() {
	// Register components used in this example.
	if err := liveflux.Register(&SingleIncludeComponent{}); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.Register(&MultiIncludeComponent{}); err != nil {
		log.Fatal(err)
	}

	handler := liveflux.NewHandler(nil)

	// Serve the demo page showcasing data-flux-include with single and multiple elements.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := hb.Webpage().
			SetTitle("Liveflux data-flux-include Example").
			SetCharset("utf-8").
			Style(`
        body { font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; max-width: 960px; margin: 40px auto; padding: 24px; background: #f9fafb; }
        h1 { font-size: 1.75rem; margin-bottom: 0.5rem; }
        p.lead { color: #4b5563; margin-bottom: 1.5rem; }
        .grid { display: grid; grid-template-columns: 1.3fr 1fr; gap: 24px; align-items: flex-start; }
        .card { background: #ffffff; border-radius: 0.75rem; padding: 20px 20px 16px; box-shadow: 0 1px 3px rgba(15,23,42,0.06); border: 1px solid #e5e7eb; }
        .card h2 { font-size: 1.1rem; margin-bottom: 0.5rem; }
        .card p { font-size: 0.9rem; color: #4b5563; }
        .toolbar { display: flex; gap: 8px; margin: 12px 0 20px; }
        .toolbar input[type="text"] { flex: 1; padding: 8px 10px; border-radius: 0.5rem; border: 1px solid #d1d5db; font-size: 0.9rem; }
        .btn { display: inline-flex; align-items: center; justify-content: center; gap: 6px; padding: 8px 12px; border-radius: 9999px; border: 1px solid #111827; background: #111827; color: #f9fafb; font-size: 0.85rem; cursor: pointer; }
        .btn.secondary { background: #ffffff; color: #111827; border-color: #d1d5db; }
        .btn:hover { filter: brightness(1.05); }
        .pill { display: inline-flex; align-items: center; gap: 6px; padding: 4px 8px; border-radius: 9999px; font-size: 0.8rem; background: #eff6ff; color: #1d4ed8; margin-right: 6px; }
        .items { margin-top: 10px; border-radius: 0.75rem; border: 1px solid #e5e7eb; background: #f9fafb; padding: 10px 12px; max-height: 220px; overflow: auto; }
        .items ul { list-style: none; padding: 0; margin: 0; }
        .items li { display: flex; align-items: center; gap: 8px; padding: 6px 0; font-size: 0.9rem; border-bottom: 1px solid #e5e7eb; }
        .items li:last-child { border-bottom: none; }
        .items small { color: #6b7280; }
        .status { margin-top: 10px; font-size: 0.85rem; color: #4b5563; }
      `).
			Child(
				hb.Div().Children([]hb.TagInterface{
					hb.H1().Text("data-flux-include demos"),
					hb.P().Class("lead").Text("These components show how to collect fields from anywhere in the DOM: a single global input and a group of checkboxes."),
					// Shared toolbar with a global search input that SingleIncludeComponent will include.
					hb.Div().Class("toolbar").Children([]hb.TagInterface{
						hb.Input().
							Attr("type", "text").
							Attr("name", "global_query").
							ID("global-query").
							Attr("placeholder", "Type a search term and click \"Apply Filter\" in the card below...").
							Attr("autocomplete", "off"),
					}),
					hb.Div().Class("grid").Children([]hb.TagInterface{
						hb.Div().Class("card").Children([]hb.TagInterface{
							hb.Div().Children([]hb.TagInterface{
								hb.H2().Text("Single input include"),
								hb.P().Text("The button in this component includes the global search input above using data-flux-include."),
							}),
							liveflux.PlaceholderByKind((&SingleIncludeComponent{}).GetKind()),
						}),
						hb.Div().Class("card").Children([]hb.TagInterface{
							hb.Div().Children([]hb.TagInterface{
								hb.H2().Text("Multiple checkbox include"),
								hb.P().Text("This component shows collecting multiple checkboxes and sending all selected IDs in one action."),
							}),
							liveflux.PlaceholderByKind((&MultiIncludeComponent{}).GetKind()),
						}),
					}),
					liveflux.Script(),
				}),
			)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write([]byte(page.ToHTML())); err != nil {
			return
		}
	})

	http.Handle("/liveflux", handler)

	fmt.Println("Server running at http://localhost:8080")
	fmt.Println("Open the page and watch Network -> Form data to see data-flux-include in action.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
