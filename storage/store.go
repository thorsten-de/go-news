package storage

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/thorsten-de/go-news/domain"
	"go.etcd.io/bbolt"
)

const articlesBucktName = "articles"

var _ domain.Storage = (*BoltStore)(nil)

type BoltStore struct {
	db *bbolt.DB
}

// NewBoltStore creates a new BoltStore instance with the specified database path and read-only mode.
func NewBoltStore(dbPath string, readOnly bool) (*BoltStore, error) {
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout:  2 * time.Second,
		ReadOnly: readOnly})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if !readOnly {
		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(articlesBucktName))
			return err
		})
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &BoltStore{db: db}, nil
}

// Closes the database connection to release database resources.
func (s *BoltStore) Close() error {
	return s.db.Close()
}

// AddArticles adds a list of articles to the database.
func (s *BoltStore) AddArticles(articles []*domain.Article) error {
	// Use an update transaction to ensure atomicity of the write operations.
	return s.db.Update(func(tx *bbolt.Tx) error {
		fmt.Println("Adding articles...")
		bucket := tx.Bucket([]byte(articlesBucktName))
		if bucket == nil {
			return fmt.Errorf("articles bucket not found")
		}
		for _, article := range articles {
			data, err := json.Marshal(article)
			if err != nil {
				return fmt.Errorf("failed to marshal article %q: %w", article.ID, err)
			}
			if err := bucket.Put([]byte(article.ID), data); err != nil {
				return fmt.Errorf("failed to store article %q: %w", article.ID, err)
			}
		}

		fmt.Println("Articles added successfully")
		return nil
	})
}

func (s *BoltStore) GetRecent(n int) []*domain.Article {
	articles := make([]*domain.Article, 0, n)
	// Use a read transaction to retrieve the most recent articles. A read transaction
	// allows for concurrent reads without blocking the database for writes. A view never
	// sees partial writes.
	s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(articlesBucktName))
		if bucket == nil {
			return nil
		}

		bucket.ForEach(func(k, v []byte) error {
			var article domain.Article
			if err := json.Unmarshal(v, &article); err != nil {
				return fmt.Errorf("failed to unmarshal article %q: %w", string(k), err)
			}
			articles = append(articles, &article)
			return nil
		})
		return nil
	})

	slices.SortFunc(articles, domain.OrderByPublished)

	if n > len(articles) {
		n = len(articles)
	}
	return articles[:n]
}

// GetByID retrieves an article by its ID. If the article is not found, it returns nil without error.
// This corresponds to a 404 status code. When an error occurs, it is treated as a 500 status code.
func (s *BoltStore) GetByID(id string) (*domain.Article, error) {
	var article *domain.Article
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(articlesBucktName))
		if bucket == nil {
			return fmt.Errorf("articles bucket not found")
		}
		data := bucket.Get([]byte(id))
		if data == nil {
			return nil // article not found, this is not treated as an error
		}
		article = &domain.Article{}
		if err := json.Unmarshal(data, article); err != nil {
			return fmt.Errorf("failed to unmarshal article: %w", err)
		}
		return nil
	})

	return article, err
}
