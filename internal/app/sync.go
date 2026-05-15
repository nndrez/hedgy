package app

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nndrez/hedgy/internal/rss"
	"github.com/nndrez/hedgy/internal/storage"
)

func (a *App) FetchAll() error {
	feeds, err := a.repo.GetFeeds()
	if err != nil {
		return fmt.Errorf("error fetching feeds: %w", err)
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
		sem  = make(chan struct{}, 5)
	)

	for _, f := range feeds {
		wg.Add(1)

		go func(feed storage.Feed) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if err := a.fetchAndSave(feed); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(f)
	}

	wg.Wait()

	return errors.Join(errs...)
}

func (a *App) FetchFeed(feed storage.Feed) error {
	return a.fetchAndSave(feed)
}

func (a *App) fetchAndSave(feed storage.Feed) error {
	entries, err := rss.Fetch(feed.URL)
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", feed.Name, err)
	}

	articles := make([]storage.Article, 0, len(entries))
	for _, entry := range entries {
		pubDate := time.Now()
		if entry.PublishedParsed != nil {
			pubDate = *entry.PublishedParsed
		}

		articles = append(articles, storage.Article{
			FeedID:      feed.ID,
			Title:       entry.Title,
			Link:        entry.Link,
			Description: entry.Description,
			Content:     entry.Content,
			PublishedAt: pubDate,
			IsRead:      false,
		})
	}

	if err := a.repo.SaveArticles(feed.ID, articles); err != nil {
		return fmt.Errorf("error saving articles for %s: %w", feed.Name, err)
	}

	return nil
}
