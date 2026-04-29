package app

import (
	"fmt"
	"sync"

	"github.com/nndrez/hedgy/internal/rss"
	"github.com/nndrez/hedgy/internal/storage"
)

type App struct {
	repo storage.Repository
}

func NewApp(repo storage.Repository) *App {
	return &App{repo: repo}
}

func (a *App) FetchAll() ([]rss.Entry, error) {
	feeds, err := a.repo.GetFeeds()
	if err != nil {
		return nil, fmt.Errorf("error fetching feeds: %w", err)
	}

	var wg sync.WaitGroup
	resultsChan := make(chan []rss.Entry, len(feeds))

	for _, f := range feeds {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()
			entries, err := rss.Fetch(url)
			if err != nil {
				return
			}

			resultsChan <- entries
		}(f.URL)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var allEntries []rss.Entry
	for entries := range resultsChan {
		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}
