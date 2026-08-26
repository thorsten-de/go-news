package domain

import "context"

// SearchResult holds the result of a search query
type SearchResult struct {
	Article *Article
	Score   float32 // Similarity score 0..1, where 1 is most similar
}

// Searchable is an interface providing search functionality
type Searchable interface {
	// Search performs a search query and returns a slice of SearchResult
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
}
