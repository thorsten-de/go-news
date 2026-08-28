package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/thorsten-de/go-news/api/internal/handlers"
	"github.com/thorsten-de/go-news/storage"
	"github.com/thorsten-de/go-news/summarizer"
)

func main() {
	dbPath := getEnv("DB_PATH", "../articles.db")
	qdrantHost := getEnv("QDRANT_HOST", "localhost")
	qdrantCollection := getEnv("QDRANT_COLLECTION", "articles")
	ollamaUrl := getEnv("OLLAMA_URL", "http://localhost:11434")
	llmModel := getEnv("LLM_MODEL", "llama3.1")
	// Open database with read-only access and defer closing when done.
	// - We enforce the principle of least privilege by opening the database
	//   in read-only mode. This provides defense in depth against unauthorized
	//   modifications to the database by the api.
	// - It also improves concurrency, as BoltDB supports multiple readers acquiring
	//   a shared lock on the database, allowing them to read simultaneously, while
	//   write mode requires an exclusive lock.
	store, err := storage.NewBoltSearcStore(dbPath, storage.Config{
		QdrantHost:       qdrantHost,
		QdrantCollection: qdrantCollection,
		OllamaURL:        ollamaUrl,
	})
	if err != nil {
		log.Fatalf("failed to initialize search store: %v", err)
	}
	defer store.Close()
	log.Printf("Storage: %v", store)

	sum, err := summarizer.New(summarizer.Config{
		OllamaURL: ollamaUrl,
		Model:     llmModel,
	})
	if err != nil {
		log.Fatalf("failed to initialize summarizer: %v", err)
	}
	log.Printf("Summarizer: model=%s", llmModel)
	// Inject storage into handlers

	// Create a new HTTP router and register route handlers
	mux := http.NewServeMux()
	handlers.NewArticleHandlers(store).RegisterRoutes(mux)
	handlers.NewSearchHandlers(store).RegisterRoutes(mux)
	handlers.NewSummaryHandlers(store, store, sum).RegisterRoutes(mux)

	// Start the HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", getEnv("PORT", "8080")),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second, // allows for the summary handler timeout to complete
	}
	fmt.Printf("Listening on :%s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// getEnv returns the value of the environment variable
// named by the key, or defaultValue if not set
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
