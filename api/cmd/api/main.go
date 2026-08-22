package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/thorsten-de/go-news/api/internal/handlers"
	"github.com/thorsten-de/go-news/storage"
)

func main() {
	// Open database access and defer closing when done
	store, err := storage.NewBoltStore("./articles.db", true)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Inject storage into handlers
	h := handlers.New(store)

	// Create a new HTTP router and register route handlers
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Start the HTTP server
	fmt.Println("Server is running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
