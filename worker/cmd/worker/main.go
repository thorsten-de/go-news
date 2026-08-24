package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thorsten-de/go-news/domain"
	"github.com/thorsten-de/go-news/storage"
	"github.com/thorsten-de/go-news/worker/internal/reader"
)

var feeds = []string{
	"https://rss.nytimes.com/services/xml/rss/nyt/World.xml",
	"https://feeds.bbci.co.uk/news/rss.xml",
}

func main() {
	// Open database in write mode, as workers persists articles.
	// Defer the close to ensure it is called when the program exits.
	// - Write mode requires an exclusive lock that is released immediately
	//   after each commit. This keeps the API available for read-only
	//   operations with minimal blocking.
	store, err := storage.NewBoltStore("./articles.db", false)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer store.Close()

	// Create an RSS reader to fetch articles from the given URLs.
	fetcher := reader.NewRSSReader()

	// Create a context with a cancel function to handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling to gracefully shutdown the worker on
	// SIGINT or SIGTERM:
	// - We use a channel to receive the signal
	// - We react by calling the cancel function to signal the context to
	//   stop all running goroutines
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutdown signal received, stopping worker...")
		cancel()
	}()

	// Fetch and store articles every 5 minutes using a ticker.
	fmt.Println("Worker started, fetching feeds every 5 minutes...")
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Fetch and store articles immediately on startup.
	fetchAndStore(ctx, fetcher, store, feeds)

	// Fetch and store articles every 5 minutes using the ticker.
	// We receive messages from
	// - the ticker channel and
	// - the context done channel
	// and coordinate our worker's behavior accordingly.
	for {
		select {
		case <-ticker.C:
			fetchAndStore(ctx, fetcher, store, feeds)
		case <-ctx.Done():
			fmt.Println("Worker stopped gracefully.")
			return
		}
	}
}

// fetchAndStore fetches and stores articles from the given URLs using the provided
// Fetcher and Storage interfaces. It skips urls that fail to fetch or store, and
// logs errors for any that do not succeed.
func fetchAndStore(ctx context.Context,
	fetcher domain.Fetcher,
	store domain.Storage,
	urls []string) {
	for _, url := range urls {
		// Use domain Fetcher interface to fetch a feed
		feed, err := fetcher.FetchFeed(ctx, url)
		if err != nil {
			log.Printf("Error fetching %s: %v", url, err)
			continue
		}
		// Use domain Storage interface to store articles
		if err := store.AddArticles(ctx, feed.Articles); err != nil {
			log.Printf("Error storing articles from %s: %v", url, err)
			continue
		}

		fmt.Printf("Fetched and stored %d articles from %s\n", len(feed.Articles), feed.Title)
	}
}
