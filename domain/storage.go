package domain

// ArticleReader provides read-only access to articles.
type ArticleReader interface {
	GetRecent(n int) []*Article
	GetByID(id string) (*Article, error)
}

// Storage persists and retrieves articles
type Storage interface {
	ArticleReader
	AddAricles(articles []*Article) error
}
