package reader

import (
	"context"

	"github.com/holmes89/gofeed"
	"github.com/thorsten-de/go-news/domain"
)

// We check at compile time that RSSReader implements the Fetcher interface
// rather than discover this at runtime
var _ domain.Fetcher = (*RSSReader)(nil)

// We implement a Fetcher that uses gofeed to parse RSS feeds
type RSSReader struct {
	parser *gofeed.Parser
}

// NewRSSReader returns a new RSSReader that uses gofeed to parse RSS feeds
func NewRSSReader() *RSSReader {
	return &RSSReader{
		parser: gofeed.NewParser(),
	}
}

// FetchFeed fetches and parses an RSS feed from the given URL.
// This method implements the Fetcher interface.
// We bridge the port-adapter boundary by translating the gofeed Feed
// and Item types into our domain types Feed and Article
func (r *RSSReader) FetchFeed(ctx context.Context, url string) (*domain.Feed, error) {
	feed, err := r.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}

	domainFeed := &domain.Feed{
		Title:       feed.Title,
		Description: feed.Description,
		Link:        feed.Link,
		Articles:    make([]*domain.Article, 0, len(feed.Items)),
	}

	for _, item := range feed.Items {
		article := &domain.Article{
			ID:          item.GUID,
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Published:   item.PublishedParsed,
			FeedTitle:   feed.Title,
		}

		domainFeed.Articles = append(domainFeed.Articles, article)
	}

	return domainFeed, nil
}
