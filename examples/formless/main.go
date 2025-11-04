package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dracory/liveflux"
)

func main() {
	// Register components
	liveflux.RegisterByAlias("formless.product-list", &ProductList{})
	liveflux.RegisterByAlias("formless.article-list", &ArticleList{})
	liveflux.RegisterByAlias("formless.multi-step", &MultiStepForm{})
	liveflux.RegisterByAlias("formless.exclude-example", &ExcludeExample{})

	// Create handler
	handler := liveflux.NewHandler(nil)

	// Serve static page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := buildPage()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page.ToHTML())
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
