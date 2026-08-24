package domain

import "context"

// Embedder is an interface for embedding text into a vector space.
type Embedder interface {
	// Embed embeds the given text into a vector space and returns the embedding.
	Embed(ctx context.Context, text string) ([]float32, error)
}
