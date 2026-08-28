package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/thorsten-de/go-news/domain"
	"github.com/thorsten-de/go-news/summarizer"
)

// SummaryHandlers handles summary requests. They depend on the storage, searcher, and summarizer.
type SummaryHandlers struct {
	storage    domain.Storage
	searcher   domain.Searchable
	summarizer *summarizer.Summarizer
}

// NewSummaryHandlers creates handler to support AI summary capabilities.
// storage and searcher can be the same store in practice.
func NewSummaryHandlers(
	storage domain.Storage,
	searcher domain.Searchable,
	summarizer *summarizer.Summarizer,
) *SummaryHandlers {
	return &SummaryHandlers{
		storage:    storage,
		searcher:   searcher,
		summarizer: summarizer,
	}
}

type SummaryResponse struct {
	Summary      string    `json:"summary"`
	ArticleCount int       `json:"article_count"`
	FocusTopic   string    `json:"focus_topic,omitempty"`
	GeneratedAt  time.Time `json:"generated_at"`
}

func (sh *SummaryHandlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /summary", sh.handleSummary)
}

func (sh *SummaryHandlers) handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompt := r.URL.Query().Get("prompt")
	count := 10
	if s := r.URL.Query().Get("count"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			count = n
		}
	}

	// Give the LLM 30 seconds to generate a summary
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var articles []*domain.Article

	if prompt != "" {
		// Query-focused path: delegate to searcher to use semantic search on the prompt to
		// retrieve relevant articles
		results, err := sh.searcher.Search(ctx, prompt, count)
		if err != nil {
			log.Printf("search error: %v", err)
			http.Error(w, "Failed to search articles", http.StatusInternalServerError)
			return
		}
		articles = make([]*domain.Article, 0, len(results))
		for _, result := range results {
			articles = append(articles, result.Article)
		}
	} else {
		// General path: retrieve the most recent articles directly from storage
		articles = sh.storage.GetRecent(count)
	}

	if len(articles) == 0 {
		http.Error(w, "No articles found to summarize", http.StatusNotFound)
		return
	}

	summary, err := sh.summarizer.Summarize(ctx, articles, prompt)
	if err != nil {
		log.Printf("summarize error: %v", err)
		http.Error(w, "Failed to summarize articles", http.StatusInternalServerError)
		return
	}
	log.Printf("summary: %d articles, topic=%q", len(articles), prompt)

	// Return the summary to the client
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(SummaryResponse{
		Summary:      summary,
		ArticleCount: len(articles),
		FocusTopic:   prompt,
		GeneratedAt:  time.Now().UTC(),
	}); err != nil {
		log.Printf("encode error: %v", err)
		http.Error(w, "Failed to encode summary", http.StatusInternalServerError)
	}
}
