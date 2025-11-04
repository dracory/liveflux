package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

func main() {
	mux := http.NewServeMux()

	// Liveflux endpoint
	mux.Handle("/liveflux", liveflux.NewHandler(nil))

	// Home page renders one Tree component and includes the client script
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tree := liveflux.SSR(&Tree{Title: "Continents and Countries"})

		page := hb.Webpage().
			SetTitle("Liveflux Tree").
			SetCharset("utf-8").
			Meta(hb.Meta().Attr("name", "viewport").Attr("content", "width=device-width, initial-scale=1")).
			StyleURLs([]string{
				"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css",
				"https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.css",
			}).
			ScriptURL("/liveflux").
			Child(
				hb.Div().Class("p-4").
					Child(
						hb.Div().Class("container").
							Children([]hb.TagInterface{
								hb.H3().Class("mb-3").Text("Tree Example"),
								tree,
							}),
					),
			)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page.ToHTML()))
	})

	port := "8085"
	addr := ":" + port
	fmt.Println("========================")
	fmt.Println("Server running at: http://localhost:" + port)
	fmt.Println("Current time: " + time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("========================")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
