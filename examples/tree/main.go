package main

import (
	"log"
	"net/http"

	"github.com/dracory/liveflux"
)

func main() {
	mux := http.NewServeMux()

	// Liveflux endpoint
	mux.Handle("/liveflux", liveflux.NewHandler(nil))

	// Home page renders one Tree component and includes the client script
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">` +
			`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css">` +
			`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.css">` +
			`<title>Liveflux Tree</title></head><body class="p-4">`

		html += `<div class="container">`
		html += `<h3 class="mb-3">Tree Example</h3>` + liveflux.SSR(&Tree{Title: "Continents and Countries"}).ToHTML()
		html += `</div>`

		// Client runtime (include once)
		html += liveflux.Script().ToHTML()

		html += "</body></html>"
		_, _ = w.Write([]byte(html))
	})

	addr := ":8080"
	log.Printf("Liveflux tree example listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
