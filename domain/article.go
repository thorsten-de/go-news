package domain

import (
	"time"
)

// Article represents a news article in an RSS feed
type Article struct {
	ID          string
	Title       string
	Description string
	Link        string
	Published   *time.Time
	FeedTitle   string
}
