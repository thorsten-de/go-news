package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/thorsten-de/go-news/domain"
)

var _ domain.SearchableStorage = (*SearchStore)(nil)

// SearchStore implements the SearchableStorage interface using Qdrant as the vector database
// It embeds the Storage interface to auto-delegate storage operations and adds
// vector search capabilities.
// Using a decorator pattern keeps storage and search concerns separated. Thus,
// we can decorate any Storage implementation with SearchStore.
type SearchStore struct {
	// Embedding the Storage interface to auto-delegate storage operations
	domain.Storage
	// Converts text to embeddings
	embedder domain.Embedder
	// Qdrant client for vector search operations
	qdrant *QdrantClient
}

// NewSearchStore creates a new SearchStore that wraps the given Storage and adds vector search capabilities.
func NewSearchStore(storage domain.Storage, embedder domain.Embedder, qdrant *QdrantClient) *SearchStore {
	return &SearchStore{
		Storage:  storage,
		embedder: embedder,
		qdrant:   qdrant,
	}
}

// NewSearchStoreWithDefaults creates a new SearchStore with Ollama embedder and Qdrant client.
func NewSearchStoreWithDefaults(ctx context.Context, storage domain.Storage) (*SearchStore, error) {
	embedder := NewOllamaEmbedder("", "")
	qdrant, err := NewQdrantClient(ctx, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant: %w", err)
	}
	return NewSearchStore(storage, embedder, qdrant), nil
}

// AddArticles adds articles to the store, embedding them as vectors if possible.
// It delegates to the embedded Storage implementation.
func (s *SearchStore) AddArticles(ctx context.Context, articles []*domain.Article) error {
	// Delegate to the embedded Storage implementation
	if err := s.Storage.AddArticles(ctx, articles); err != nil {
		return err
	}

	for _, article := range articles {
		if err := s.embedAndStore(ctx, article); err != nil {
			log.Printf("Warning: Failed to embed article %s: %v", article.ID, err)
			// Continue with the next article, as we can still store non-embeddable articles
		}
	}

	return nil
}

func (s *SearchStore) embedAndStore(ctx context.Context, article *domain.Article) error {
	// Combine title and description to embed a richer semantic meaning
	text := article.Title
	if article.Description != "" {
		text += "\n\n" + article.Description
	}

	// Generate vector embedding
	vector, err := s.embedder.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to embed text: %w", err)
	}

	// Store the embedding in the vector database with metadata for filtering
	metadata := map[string]any{
		"title":      article.Title,
		"feed_title": article.FeedTitle,
	}

	return s.qdrant.Store(ctx, article.ID, vector, metadata)
}

// Search performs a semantic search on the vector database using the given query. It returns articles
// sorted by similarity to the query.
func (s *SearchStore) Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	// Convert query text to embedding vector
	queryVector, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Find similar articles in the vector database
	matches, err := s.qdrant.Search(ctx, queryVector, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar articles: %w", err)
	}

	// Fetch the full article details for each match from the underlying storage
	results := make([]domain.SearchResult, 0, len(matches))
	for _, match := range matches {
		article, err := s.Storage.GetByID(match.ID)
		if err != nil {
			// Log the error and skip this match
			log.Printf("Warning: Failed to fetch article: %s: %v", match.ID, err)
			continue
		}
		results = append(results, domain.SearchResult{
			Article: article,
			Score:   match.Score,
		})
	}

	return results, nil
}
