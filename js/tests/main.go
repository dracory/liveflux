package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// contentTypeMiddleware sets the correct Content-Type header for JavaScript files
func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := strings.ToLower(filepath.Ext(r.URL.Path))
		switch ext {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case ".json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := "8000"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Serve from parent directory (js/) to access all JS files
	parentDir := filepath.Join(dir, "..")
	
	// Serve files from parent directory with correct MIME types
	fs := http.FileServer(http.Dir(parentDir))
	http.Handle("/", contentTypeMiddleware(fs))

	fmt.Printf("Liveflux JS Test Server\n")
	fmt.Printf("========================\n")
	fmt.Printf("Server running at: http://localhost:%s\n", port)
	fmt.Printf("Test runner: http://localhost:%s/tests/runner.html\n", port)
	fmt.Printf("Serving files from: %s\n", parentDir)
	fmt.Printf("\nPress Ctrl+C to stop the server\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
