package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thorsten-de/go-news/api/internal/handlers"
	"github.com/thorsten-de/go-news/domain"
)

// mockReader implements domain.ArticleReader for testing.
type mockReader struct {
	articles []*domain.Article
}

func (m *mockReader) GetRecent(n int) []*domain.Article {
	if n > len(m.articles) {
		n = len(m.articles)
	}
	return m.articles[:n]
}

func (m *mockReader) GetByID(id string) (*domain.Article, error) {
	// not needed for these tests
	return nil, nil
}

// newMockReader creates a mockReader populated with `count` dummy articles.
func newMockReader(count int) *mockReader {
	articles := make([]*domain.Article, count)
	for i := 0; i < count; i++ {
		// Use a simple ID and Title for each article.
		articles[i] = &domain.Article{ID: fmt.Sprintf("id%d", i), Title: fmt.Sprintf("Title %d", i)}
	}
	return &mockReader{articles: articles}
}

func TestArticlesHandler_DefaultCount(t *testing.T) {
	// Prepare 15 dummy articles via helper.
	h := handlers.New(newMockReader(15))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/articles", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}
	var resp []*domain.Article
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	// Default should be 10 articles.
	if len(resp) != 10 {
		t.Fatalf("expected 10 articles, got %d", len(resp))
	}
}

func TestArticlesHandler_CustomCount(t *testing.T) {
	h := handlers.New(newMockReader(8))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/articles?count=5", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var resp []*domain.Article
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 5 {
		t.Fatalf("expected 5 articles, got %d", len(resp))
	}
}

func TestArticlesHandler_MethodNotAllowed(t *testing.T) {
	h := handlers.New(newMockReader(0))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/articles", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
