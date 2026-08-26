package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/thorsten-de/go-news/domain"
)

// Handlers manage HTTP requests handlers with their dependencies.
type Handlers struct {
	// Get read-only access to articles. We use an interface to allow
	// mocking our storage implementation and decouple the handler from it.
	// This respects both ISP and LSP.
	articles domain.ArticleReader
}

// NewArticleHandlers creates a new Handlers instance with the given article reader.
func NewArticleHandlers(articles domain.ArticleReader) *Handlers {
	return &Handlers{
		articles: articles,
	}
}

// RegisterRoutes mounts all handler routes on the provided mux.
// We can provide the mux, so we can replace it with a test double.
func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	// GET /articles - returns the recent articles
	mux.HandleFunc("/articles", h.handleArticles)
}

// Request handler returning the recent articles.
func (h *Handlers) handleArticles(w http.ResponseWriter, r *http.Request) {
	// handle GET requests only
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	n := 10
	if countStr := r.URL.Query().Get("count"); countStr != "" {
		if count, err := strconv.Atoi(countStr); err == nil {
			n = count
		}
	}

	articles := h.articles.GetRecent(n)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(articles); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
