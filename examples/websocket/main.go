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

	mux.Handle("/liveflux", liveflux.NewHandlerWS(nil))

	// Home page renders the SSR for two WebSocketCounters
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create two instances
		inst1, err := liveflux.New(&WebSocketCounter{})
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		inst2, err := liveflux.New(&WebSocketCounter{})
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		html := `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">` +
			`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css">` +
			`<title>Liveflux WebSocket Counter</title></head><body class="p-4">`

		// SSR two components side-by-side
		html += `<div class="container"><div class="row g-4">`
		html += `<div class="col-md-6"><h3>WebSocket Counter 1</h3>` + liveflux.SSR(inst1).ToHTML() + `</div>`
		html += `<div class="col-md-6"><h3>WebSocket Counter 2</h3>` + liveflux.SSR(inst2).ToHTML() + `</div>`
		html += `</div></div>`

		// Include WebSocket client (from core js/websocket.js)
		html += `<script src="/liveflux"></script>`
		// Inline debug bootstrap to verify client is loaded and elements are found
		html += `<script>
          console.log('[BOOT] inline script running');
          console.log('[BOOT] LiveFluxWS typeof =', typeof window.LiveFluxWS);
          const els = document.querySelectorAll('[data-flux-ws]');
          console.log('[BOOT] ws elements found =', els.length, els);
        </script>`

		html += "</body></html>"
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(html))
	})

	port := "8091"
	addr := ":" + port
	fmt.Printf("========================\n")
	fmt.Printf("Server running at: http://localhost:%s\n", port)
	fmt.Println("Current time: " + time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("========================\n")
	log.Printf("Liveflux WebSocket counter example listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
