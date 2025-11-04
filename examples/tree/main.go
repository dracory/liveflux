package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

	port := "8085"
	addr := ":" + port
	fmt.Printf("========================\n")
	fmt.Printf("Server running at: http://localhost:%s\n", port)
	fmt.Println("Current time: " + time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("========================\n")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
