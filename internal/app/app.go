package app

import (
	"fmt"
	"sync"
	"time"

	"github.com/nndrez/hedgy/internal/rss"
	"github.com/nndrez/hedgy/internal/storage"
)

type App struct {
	repo storage.Repository
}

func NewApp(repo storage.Repository) *App {
	return &App{repo: repo}
}

func (a *App) FetchAll() error {
	feeds, err := a.repo.GetFeeds()
	if err != nil {
		return fmt.Errorf("error fetching feeds: %w", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(feeds))

	for _, f := range feeds {
		wg.Add(1)

		go func(feed storage.Feed) {
			defer wg.Done()
			entries, err := rss.Fetch(feed.URL)
			if err != nil {
				errChan <- fmt.Errorf("error downloading %s: %w", feed.Name, err)
				return
			}

			var articles []storage.Article
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

			err = a.repo.SaveArticles(feed.ID, articles)
			if err != nil {
				errChan <- fmt.Errorf("error saving articles fot %s: %w", feed.Name, err)
				return
			}
		}(f)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
