package main

import (
	"log"
	"net/http"

	"github.com/dracory/liveflux"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/liveflux", liveflux.NewHandler(nil))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		list := liveflux.SSR(&UserList{}).ToHTML()
		create := liveflux.SSR(&CreateUserModal{}).ToHTML()
		edit := liveflux.SSR(&EditUserModal{}).ToHTML()
		deleteHTML := liveflux.SSR(&DeleteUserModal{}).ToHTML()

		html := `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">` +
			`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css">` +
			`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.css">` +
			`<title>Liveflux CRUD</title>` +
			`<style>` +
			`body{background-color:#f8f9fa;}` +
			`.crud-modal{display:none;position:fixed;inset:0;padding:1.5rem;background:rgba(33,37,41,.55);z-index:1050;align-items:center;justify-content:center;}` +
			`.crud-modal__card{background:#fff;border-radius:.9rem;box-shadow:0 2rem 4rem rgba(15,23,42,.3);overflow:hidden;max-width:520px;width:100%;}` +
			`.crud-modal__header{padding:1rem 1.5rem;border-bottom:1px solid rgba(0,0,0,.1);display:flex;align-items:center;justify-content:space-between;gap:1rem;}` +
			`.crud-modal__body{padding:1.5rem;display:flex;flex-direction:column;gap:1rem;}` +
			`.crud-modal__footer{padding:0 1.5rem 1.5rem 1.5rem;display:flex;gap:.75rem;justify-content:flex-end;}` +
			`.crud-badge{font-size:.75rem;letter-spacing:.04em;text-transform:uppercase;}` +
			`</style></head><body>`

		html += `<div class="container py-4"><div class="mb-4 text-center"><h1 class="fw-semibold">Team Directory</h1>` +
			`<p class="text-muted mb-0">Manage a simple in-memory list of teammates with Liveflux.</p></div>`
		html += list + create + edit + deleteHTML + `</div>`
		html += liveflux.Script().ToHTML()
		html += `</body></html>`

		_, _ = w.Write([]byte(html))
	})

	addr := ":8080"
	log.Printf("Liveflux CRUD example listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
