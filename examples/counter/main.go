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

	// Home page renders the SSR for two Counters and includes the client script
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create two instances from the registry
		inst1, err := liveflux.New(&Counter{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error: " + err.Error()))
			return
		}
		inst2, err := liveflux.New(&Counter{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error: " + err.Error()))
			return
		}

		html := `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">` +
			`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css">` +
			`<title>Liveflux Counter</title></head><body class="p-4">`

		// SSR two components side-by-side
		html += `<div class="container"><div class="row g-4">`
		html += `<div class="col-md-6"><h3>Counter 1</h3>` + liveflux.SSR(inst1).ToHTML() + `</div>`
		html += `<div class="col-md-6"><h3>Counter 2</h3>` + liveflux.SSR(inst2).ToHTML() + `</div>`
		html += `</div></div>`

		// Client runtime (include once)
		html += liveflux.Script().ToHTML()

		html += "</body></html>"
		_, _ = w.Write([]byte(html))
	})

	addr := ":8080"
	log.Printf("Liveflux counter example listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
