package domain

import (
	"context"
)

// A fetcher retrieves and parses an RSS feed from a given URL.
type Fetcher interface {
	FetchFeed(ctx context.Context, url string) (*Feed, error)
}
