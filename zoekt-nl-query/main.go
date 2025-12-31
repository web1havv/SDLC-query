package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"github.com/sourcegraph/zoekt/search"
)

func main() {
	indexDir := flag.String("index", filepath.Join("/Users/web1havv", ".zoekt"), "index directory")
	port := flag.String("port", "6071", "port to serve on")
	flag.Parse()

	log.Printf("Starting NL Query Server on port %s", *port)
	log.Printf("Index directory: %s", *indexDir)

	// Create searcher
	searcher, err := search.NewDirectorySearcher(*indexDir)
	if err != nil {
		log.Fatalf("Failed to create searcher: %v", err)
	}

	// Create NL query server
	nlServer := NewNLQueryServer(searcher)

	// Setup routes
	mux := http.NewServeMux()
	nlServer.SetupRoutes(mux)

	// Copy dashboard.html to current directory for serving
	log.Println("Starting server...")
	log.Printf("Dashboard: http://localhost:%s/dashboard", *port)
	log.Printf("NL Search API: http://localhost:%s/api/nl-search?q=how many articles", *port)
	log.Printf("Index directory: %s", *indexDir)

	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

