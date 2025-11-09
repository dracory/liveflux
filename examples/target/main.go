package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

func main() {
	// Register components used in this example.
	if err := liveflux.Register(&CartComponent{}); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.Register(&DealsComponent{}); err != nil {
		log.Fatal(err)
	}

	handler := liveflux.NewHandler(nil)

	// Serve the demo page showcasing multiple components and a document-scoped badge.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// Render the static badge once so the outer DOM has content before Liveflux mounts components.
		initialCart := &CartComponent{}
		if err := initialCart.Mount(ctx, nil); err != nil {
			log.Printf("failed to mount initial cart: %v", err)
		}

		page := hb.Webpage().
			SetTitle("Liveflux Target Example - Shopping Cart").
			SetCharset("utf-8").
			Style(`
                body { font-family: Arial, sans-serif; max-width: 900px; margin: 40px auto; padding: 20px; }
                .top-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 30px; padding: 15px 20px; background: #0f172a; color: #fff; border-radius: 8px; }
                .cart-badge { font-weight: 600; }
                .grid { display: grid; grid-template-columns: 2fr 1fr; gap: 20px; }
                .card { border: 1px solid #ddd; border-radius: 8px; padding: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.04); background: #fff; }
                .cart-container { border: none; box-shadow: none; padding: 0; }
                .deal-banner { padding: 12px; background: #e0f2fe; border-radius: 6px; margin-bottom: 15px; font-weight: 600; }
                .btn { padding: 10px 16px; cursor: pointer; border-radius: 6px; border: 1px solid #0f172a; background: #0f172a; color: #fff; }
                .btn:hover { background: #1e293b; }
            `).
			Child(
				hb.Div().Children([]hb.TagInterface{
					hb.Div().Class("top-bar").Children([]hb.TagInterface{
						hb.H1().Text("Liveflux Targeted Updates Demo"),
						// Standalone DOM element outside any component root; updated via TargetScopeDocument.
						initialCart.renderGlobalBadge(),
					}),
					hb.Div().Class("grid").Children([]hb.TagInterface{
						hb.Div().Class("card cart-container").Children([]hb.TagInterface{
							hb.H2().Text("Shopping Cart"),
							hb.P().Text("Adding or removing items only updates the relevant fragments."),
							liveflux.PlaceholderByKind("cart"),
						}),
						hb.Div().Class("card").Children([]hb.TagInterface{
							hb.H2().Text("Deals Component"),
							hb.P().Text("This separate component also uses targets to update its banner."),
							liveflux.PlaceholderByKind("deals"),
						}),
					}),
					liveflux.Script(),
				}),
			)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page.ToHTML()))
	})

	http.Handle("/liveflux", handler)

	fmt.Println("Server running at http://localhost:8080")
	fmt.Println("Open your browser and check the Network tab to see targeted updates in action!")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
