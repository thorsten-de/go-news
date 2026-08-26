package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/thorsten-de/go-news/domain"
)

// SearchHandlers handles search requests for articles. These depend on the
// searchable interface to perform the search.
type SearchHandlers struct {
	searchable domain.Searchable
}

// NewSearchHandlers creates a new SearchHandlers instance.
func NewSearchHandlers(searchable domain.Searchable) *SearchHandlers {
	return &SearchHandlers{searchable: searchable}
}

func (h *SearchHandlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/search", h.handleSearch)
}

// Search handles search requests for articles.
func (sh *SearchHandlers) handleSearch(w http.ResponseWriter, r *http.Request) {
	// Parse query and limit parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// Perform search
	results, err := sh.searchable.Search(r.Context(), query, limit)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	// Return query and results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"query":   query,
		"results": results,
	})
}
