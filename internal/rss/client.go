package rss

import (
	"context"
	"time"

	"github.com/mmcdole/gofeed"
)

type Entry struct {
	Title           string
	Link            string
	Description     string
	Content         string
	PublishedParsed *time.Time
}

const fetchTimeout = 10 * time.Second

func Fetch(url string) ([]Entry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, item := range feed.Items {
		entries = append(entries, Entry{
			Title:           item.Title,
			Link:            item.Link,
			Description:     item.Description,
			Content:         item.Content,
			PublishedParsed: item.PublishedParsed,
		})
	}

	return entries, nil
}
