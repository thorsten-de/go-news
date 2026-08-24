package domain

import "context"

// ArticleReader provides read-only access to articles.
type ArticleReader interface {
	GetRecent(n int) []*Article
	GetByID(id string) (*Article, error)
}

// Storage persists and retrieves articles
// It provides methods for reading and writing articles to a persistent storage
type Storage interface {
	ArticleReader
	AddArticles(ctx context.Context, articles []*Article) error
	Close() error
}

// SearchableStorage is a storage that supports search queries. We combine
// Storage and Searchable interfaces to provide both persistence and semantic
// search capabilities.
type SearchableStorage interface {
	Storage
	Searchable
}
