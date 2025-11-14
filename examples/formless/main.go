package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dracory/liveflux"
)

func main() {
	// Register components
	if err := liveflux.RegisterByKind("formless.product-list", &ProductList{}); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.RegisterByKind("formless.article-list", &ArticleList{}); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.RegisterByKind("formless.multi-step", &MultiStepForm{}); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.RegisterByKind("formless.exclude-example", &ExcludeExample{}); err != nil {
		log.Fatal(err)
	}

	// Create handler
	handler := liveflux.NewHandler(nil)

	// Serve static page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := buildPage()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := fmt.Fprint(w, page.ToHTML()); err != nil {
			// If we can't write the response there is nothing reasonable to do here.
			return
		}
	})

	// Mount Liveflux handler
	http.Handle("/liveflux", handler)

	port := "8080"
	addr := ":" + port
	fmt.Printf("========================\n")
	fmt.Printf("Server running at: http://localhost:%s\n", port)
	fmt.Printf("========================\n")
	log.Fatal(http.ListenAndServe(addr, nil))
}
