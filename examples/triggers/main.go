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

	// Home page renders the SSR for the search component
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		inst, err := liveflux.New(&SearchComponent{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error: " + err.Error()))
			return
		}

		searchComponent := liveflux.SSR(inst)

		page := hb.Webpage().
			SetTitle("Liveflux Triggers - Live Search").
			SetCharset("utf-8").
			Meta(hb.Meta().Attr("name", "viewport").Attr("content", "width=device-width, initial-scale=1")).
			StyleURL("https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css").
			ScriptURL("/liveflux"). // external script
			Child(
				hb.Div().Class("container py-5").
					Children([]hb.TagInterface{
						hb.Div().Class("row justify-content-center").
							Child(hb.Div().Class("col-lg-8").
								Children([]hb.TagInterface{
									hb.H1().Class("mb-4").Text("Live Search with Triggers"),
									hb.P().Class("lead mb-4").
										Text("This example demonstrates the data-flux-trigger feature. " +
											"The search input uses 'keyup changed delay:300ms' to trigger " +
											"searches as you type with automatic debouncing."),
									searchComponent,
									hb.Div().Class("mt-4").
										Children([]hb.TagInterface{
											hb.H5().Text("How it works:"),
											hb.Ul().
												Children([]hb.TagInterface{
													hb.Li().HTML("<code>keyup</code> - Triggers on every keystroke"),
													hb.Li().HTML("<code>changed</code> - Only fires if the value actually changed"),
													hb.Li().HTML("<code>delay:300ms</code> - Debounces requests by 300ms"),
													hb.Li().Text("No custom JavaScript required!"),
												}),
										}),
								})),
					}),
			)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page.ToHTML()))
	})

	port := "8082"
	addr := ":" + port
	fmt.Printf("========================\n")
	fmt.Printf("Liveflux Triggers Example\n")
	fmt.Printf("Server running at: http://localhost:%s\n", port)
	fmt.Println("Current time: " + time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("========================\n")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func init() {
	// Register the search component
	if err := liveflux.Register(new(SearchComponent)); err != nil {
		log.Fatal(err)
	}
}
