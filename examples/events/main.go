package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

func main() {
	// Register components
	if err := liveflux.RegisterByAlias("post-creator", &PostCreator{}); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.RegisterByAlias("post-list", &PostList{}); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.RegisterByAlias("notification-banner", &NotificationBanner{}); err != nil {
		log.Fatal(err)
	}

	// Setup HTTP server
	mux := http.NewServeMux()

	// Liveflux handler
	mux.Handle("/liveflux", liveflux.NewHandler(nil))

	// Main page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := hb.Webpage().
			SetTitle("Liveflux Events Example").
			SetCharset("utf-8").
			ScriptURL("/liveflux"). // external script
			// Script(liveflux.JS()). // embedded script
			Style(`
				body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; background: #f5f5f5; }
				.container { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px; }
					.card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
					.notification-card { grid-column: 1 / -1; }
					h2 { margin-top: 0; color: #333; }
					.input { width: 100%; padding: 10px; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
					.btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; }
					.btn-primary { background: #007bff; color: white; }
					.btn-primary:hover { background: #0056b3; }
					.btn-secondary { background: #6c757d; color: white; margin-top: 10px; }
					.btn-secondary:hover { background: #545b62; }
					.post-list { list-style: none; padding: 0; margin: 15px 0; }
					.post-item { padding: 10px; margin: 5px 0; background: #f8f9fa; border-radius: 4px; }
					.timestamp { color: #666; font-size: 0.9em; }
					.empty { color: #999; font-style: italic; }
					.notification { padding: 15px; background: #e9ecef; border-radius: 4px; text-align: center; }
					.notification.active { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
					form { margin: 0; }
				`).
			Children([]hb.TagInterface{
				hb.H1().Text("Liveflux Events Example"),
				hb.P().Text("This example demonstrates the event system. Create a post in the left component, and watch it appear in the list on the right via events."),
				hb.Div().Class("container").Children([]hb.TagInterface{
					liveflux.PlaceholderByAlias("post-creator"),
					liveflux.PlaceholderByAlias("post-list"),
				}),
				liveflux.PlaceholderByAlias("notification-banner"),
			})

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page.ToHTML())
	})

	addr := ":8084"
	fmt.Printf("Server running at http://localhost%s\n", addr)
	fmt.Println("Open your browser and create posts to see events in action!")
	log.Fatal(http.ListenAndServe(addr, mux))
}
