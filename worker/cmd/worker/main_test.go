package main

import (
	"context"
	"errors"
	"testing"

	"github.com/thorsten-de/go-news/domain"
)

type mockFetcher struct {
	feeds  map[string]*domain.Feed
	errors map[string]error
}

func (m *mockFetcher) FetchFeed(ctx context.Context, url string) (*domain.Feed, error) {
	if err, ok := m.errors[url]; ok {
		return nil, err
	}
	if f, ok := m.feeds[url]; ok {
		return f, nil
	}
	return nil, errors.New("not found")
}

type mockStorage struct {
	added [][]*domain.Article
	err   error
}

func (m *mockStorage) GetRecent(n int) []*domain.Article          { return nil }
func (m *mockStorage) GetByID(id string) (*domain.Article, error) { return nil, nil }
func (m *mockStorage) AddArticles(articles []*domain.Article) error {
	m.added = append(m.added, articles)
	return m.err
}

func TestFetchAndStore_AllSuccess(t *testing.T) {
	ctx := context.Background()
	urls := []string{"u1", "u2"}
	mf := &mockFetcher{feeds: map[string]*domain.Feed{
		"u1": {Title: "Feed1", Articles: []*domain.Article{{ID: "a1"}}},
		"u2": {Title: "Feed2", Articles: []*domain.Article{{ID: "a2"}}},
	}}
	ms := &mockStorage{}
	fetchAndStore(ctx, mf, ms, urls)
	if len(ms.added) != 2 {
		t.Fatalf("expected 2 calls to AddArticles, got %d", len(ms.added))
	}
	if len(ms.added[0]) != 1 || ms.added[0][0].ID != "a1" {
		t.Fatalf("unexpected first stored articles: %+v", ms.added[0])
	}
	if len(ms.added[1]) != 1 || ms.added[1][0].ID != "a2" {
		t.Fatalf("unexpected second stored articles: %+v", ms.added[1])
	}
}

func TestFetchAndStore_FetchError(t *testing.T) {
	ctx := context.Background()
	urls := []string{"good", "bad", "also-good"}
	mf := &mockFetcher{
		feeds:  map[string]*domain.Feed{"good": {Title: "Good", Articles: []*domain.Article{{ID: "g1"}}}, "also-good": {Title: "AlsoGood", Articles: []*domain.Article{{ID: "g2"}}}},
		errors: map[string]error{"bad": errors.New("fetch failed")},
	}
	ms := &mockStorage{}
	fetchAndStore(ctx, mf, ms, urls)
	if len(ms.added) != 2 {
		t.Fatalf("expected 2 successful stores, got %d", len(ms.added))
	}
}

func TestFetchAndStore_StorageError(t *testing.T) {
	ctx := context.Background()
	urls := []string{"first", "second"}
	mf := &mockFetcher{feeds: map[string]*domain.Feed{
		"first":  {Title: "First", Articles: []*domain.Article{{ID: "f1"}}},
		"second": {Title: "Second", Articles: []*domain.Article{{ID: "f2"}}},
	}}
	ms := &mockStorage{err: errors.New("store failed")}
	fetchAndStore(ctx, mf, ms, urls)
	// Even though storage returns an error, it should still be called for each URL.
	if len(ms.added) != 2 {
		t.Fatalf("expected AddArticles called for both URLs despite error, got %d", len(ms.added))
	}
}
