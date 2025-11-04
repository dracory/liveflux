package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

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

	// Serve files from current directory
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	fmt.Printf("Liveflux JS Test Server\n")
	fmt.Printf("========================\n")
	fmt.Printf("Server running at: http://localhost:%s\n", port)
	fmt.Printf("Test runner: http://localhost:%s/runner.html\n", port)
	fmt.Printf("Serving files from: %s\n", dir)
	fmt.Printf("\nPress Ctrl+C to stop the server\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
