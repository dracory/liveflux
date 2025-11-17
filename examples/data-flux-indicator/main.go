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
		demoInst, err := liveflux.New(&fetchDataComponent{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error: " + err.Error()))
			return
		}
		formInst, err := liveflux.New(&IndicatorForm{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error: " + err.Error()))
			return
		}
		externalInst, err := liveflux.New(&ExternalIndicatorDemo{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error: " + err.Error()))
			return
		}

		demoComponent := liveflux.SSR(demoInst)
		formComponent := liveflux.SSR(formInst)
		externalComponent := liveflux.SSR(externalInst)

		globalIndicator := hb.Div().
			Attr("id", "global-indicator").
			Class("alert alert-info align-items-center gap-2 mb-4").
			Style("display: none !important;").
			Child(hb.Div().Class("d-inline-flex align-items-center gap-2").
				Child(hb.Div().Class("spinner-border spinner-border-sm").Attr("role", "status").
					Child(hb.Span().Class("visually-hidden").Text("Loading"))).
				Child(hb.Span().Text("Global indicator")))

		fetchDataButton := hb.Button().
			Class("btn btn-outline-secondary mb-4").
			Attr(liveflux.DataFluxAction, "fetch").
			Attr(liveflux.DataFluxTargetKind, demoInst.GetKind()).
			Attr(liveflux.DataFluxTargetID, demoInst.GetID()).
			Attr(liveflux.DataFluxIndicator, "#global-indicator").
			Attr(liveflux.DataFluxSelect, "#status-text").
			Text("External fetch button (using global indicator)")

		fetchDataButton2 := hb.Button().
			Class("btn btn-outline-secondary mb-4").
			Attr(liveflux.DataFluxAction, "fetch").
			Attr(liveflux.DataFluxTargetKind, demoInst.GetKind()).
			Attr(liveflux.DataFluxTargetID, demoInst.GetID()).
			Attr(liveflux.DataFluxIndicator, "this, .spinner").
			Attr(liveflux.DataFluxSelect, "#status-text").
			Text("External fetch button (using local indicator)").
			Child(hb.Span().
				Class("spinner spinner-border spinner-border-sm align-middle ms-2").
				Attr("role", "status").
				Style(`display: none;`).
				Child(hb.Span().Class("visually-hidden").Text("Loading")),
			)

		page := hb.Webpage().
			SetTitle("Liveflux Indicators").
			SetCharset("utf-8").
			Meta(hb.Meta().Attr("name", "viewport").Attr("content", "width=device-width, initial-scale=1")).
			StyleURL("https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css").
			ScriptURL("/liveflux").
			Child(
				hb.Div().Class("container py-5").
					Child(hb.Div().Class("row justify-content-center").
						Child(hb.Div().Class("col-lg-10").
							Child(globalIndicator).
							Child(hb.Div().Class("row g-4").
								Child(hb.Div().
									Class("col-md-6").
									Child(demoComponent).
									Child(hb.HR()).
									Child(fetchDataButton).
									Child(hb.HR()).
									Child(fetchDataButton2).
									Child(hb.HR()),
								).
								Child(hb.Div().
									Class("col-md-6").
									Child(formComponent)).
								Child(hb.Div().
									Class("col-12").
									Child(externalComponent))))),
			)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page.ToHTML()))
	})

	port := "8084"
	addr := ":" + port
	fmt.Printf("========================\n")
	fmt.Printf("Server running at: http://localhost:%s\n", port)
	fmt.Println("Current time: " + time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("========================\n")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
