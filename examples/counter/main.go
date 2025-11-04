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

		counter1 := liveflux.SSR(inst1)
		counter2 := liveflux.SSR(inst2)

		page := hb.Webpage().
			SetTitle("Liveflux Counter").
			SetCharset("utf-8").
			Meta(hb.Meta().Attr("name", "viewport").Attr("content", "width=device-width, initial-scale=1")).
			StyleURL("https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css").
			ScriptURL("/liveflux"). // external script
			// Script(liveflux.JS()). // embedded script
			Child(
				hb.Div().Class("container py-4").
					Child(hb.Div().Class("row g-4").
						Children([]hb.TagInterface{
							hb.Div().Class("col-md-6").
								Children([]hb.TagInterface{
									hb.H3().Text("Counter 1"),
									counter1,
								}),
							hb.Div().Class("col-md-6").
								Children([]hb.TagInterface{
									hb.H3().Text("Counter 2"),
									counter2,
								}),
						})),
			)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page.ToHTML()))
	})

	port := "8081"
	addr := ":" + port
	fmt.Printf("========================\n")
	fmt.Printf("Server running at: http://localhost:%s\n", port)
	fmt.Println("Current time: " + time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("========================\n")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
