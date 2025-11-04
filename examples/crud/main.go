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

	mux.Handle("/liveflux", liveflux.NewHandler(nil))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		modalCreate := &CreateUserModal{}
		modalEdit := &EditUserModal{}
		modalDelete := &DeleteUserModal{}

		list := liveflux.SSR(&UserList{
			ModalCreateUser: modalCreate,
		})
		create := liveflux.SSR(modalCreate)
		edit := liveflux.SSR(modalEdit)
		deleteHTML := liveflux.SSR(modalDelete)

		page := hb.Webpage().
			SetTitle("Liveflux CRUD").
			SetCharset("utf-8").
			Style(`
				body{background-color:#f8f9fa;}
				.crud-modal{display:none;position:fixed;inset:0;padding:1.5rem;background:rgba(33,37,41,.55);z-index:1050;align-items:center;justify-content:center;}
				.crud-modal__card{background:#fff;border-radius:.9rem;box-shadow:0 2rem 4rem rgba(15,23,42,.3);overflow:hidden;max-width:520px;width:100%;}
				.crud-modal__header{padding:1rem 1.5rem;border-bottom:1px solid rgba(0,0,0,.1);display:flex;align-items:center;justify-content:space-between;gap:1rem;}
				.crud-modal__body{padding:1.5rem;display:flex;flex-direction:column;gap:1rem;}
				.crud-modal__footer{padding:0 1.5rem 1.5rem 1.5rem;display:flex;gap:.75rem;justify-content:flex-end;}
				.crud-badge{font-size:.75rem;letter-spacing:.04em;text-transform:uppercase;}
			`).
			Child(
				hb.Div().Class("container").
					Children([]hb.TagInterface{
						hb.H1().Text("Team Directory"),
						hb.P().Text("Manage a simple in-memory list of teammates with Liveflux."),
						list,
						create,
						edit,
						deleteHTML,
					}),
			).
			StyleURLs([]string{
				"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css",
				"https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css",
			}).
			ScriptURLs([]string{
				"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js",
				"/liveflux",
			})
			// .
			// Script(liveflux.JS())

		_, _ = w.Write([]byte(page.ToHTML()))
	})

	port := "8082"
	addr := ":" + port
	fmt.Printf("========================\n")
	fmt.Printf("Server running at: http://localhost:%s\n", port)
	fmt.Println("Open your browser and create posts to see CRUD in action!")
	fmt.Println("Current time: " + time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("========================\n")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
