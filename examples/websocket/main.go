package main

import (
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/dracory/liveflux"
)

func main() {
	mux := http.NewServeMux()

	// Liveflux WebSocket endpoint
	handler := liveflux.NewWebSocketHandler(nil)
	handler.Handle("websocket-counter", func() liveflux.Component {
		return &WebSocketCounter{}
	})
	mux.Handle("/liveflux", handler)

	// Ensure correct JS MIME type to satisfy browsers with nosniff
	_ = mime.AddExtensionType(".js", "application/javascript")

	// Resolve static directory to support running from repo root and examples/websocket
	resolveStaticDir := func() string {
		candidates := []string{
			"./static",          // when running from repo root
			"../../static",      // when running from examples/websocket
		}
		for _, dir := range candidates {
			if f, err := os.Stat(dir); err == nil && f.IsDir() {
				return dir
			}
		}
		// fallback to first
		return "./static"
	}
	staticDir := resolveStaticDir()

	// Resolve core websocket client location (repo js/websocket.js)
	resolveWSClient := func() string {
		candidates := []string{
			"./js/websocket.js",        // when running from repo root
			"../../js/websocket.js",    // when running from examples/websocket
		}
		for _, p := range candidates {
			if f, err := os.Stat(p); err == nil && !f.IsDir() {
				return p
			}
		}
		// fallback path (may 404 if not present)
		return "./js/websocket.js"
	}
	wsClientPath := resolveWSClient()

	// Serve the WS client with explicit MIME type
	mux.HandleFunc("/static/websocket.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		http.ServeFile(w, r, wsClientPath)
	})

	// Generic static files (images, css, etc.)
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		}
		fs.ServeHTTP(w, r)
	})))

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
		html += `<script src="/static/websocket.js"></script>`
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

	addr := ":8080"
	log.Printf("Liveflux WebSocket counter example listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
