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

// OrderByPublished orders articles by their published date, with nil dates coming last
func OrderByPublished(a, b *Article) int {
	if a.Published == nil && b.Published == nil {
		return 0 // don't care
	}
	if a.Published == nil {
		return 1 // push a to the end
	}
	if b.Published == nil {
		return -1 // push b to the end
	}
	return b.Published.Compare(*a.Published)
}
