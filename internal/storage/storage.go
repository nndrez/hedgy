package storage

import (
	"time"
)

type Feed struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Article struct {
	ID          int       `json:"id"`
	FeedID      int       `json:"feedId"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	PublishedAt time.Time `json:"publishedAt"`
	IsRead      bool      `json:"isRead"`
}

type Repository interface {
	GetFeeds() ([]Feed, error)
	AddFeed(name, url string) error
	DeleteFeed(id int) error

	SaveArticles(feedID int, articles []Article) error
	GetUnreadArticles(feedID int, limit int) ([]Article, error)
	ToggleReadStatus(articleID int) error
}
