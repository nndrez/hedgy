package main

import (
	"fmt"
	"log"

	"github.com/nndrez/hedgy/internal/app"
	"github.com/nndrez/hedgy/internal/storage"
)

func main() {
	repo, err := storage.NewSQLiteStorage()
	if err != nil {
		log.Fatalf("Error during storage initialization: %v", err)
	}

	feeds, err := repo.GetFeeds()
	if err != nil {
		log.Fatalf("Error during feeds reading: %v", err)
	}

	if len(feeds) == 0 {
		fmt.Println("Not found any feed. Insert test feeds...")
		repo.AddFeed("Hacker News", "https://news.ycombinator.com/rss")
		repo.AddFeed("Go Blog", "https://go.dev/blog/feed.atom")
		repo.AddFeed("Reddit Go", "https://www.reddit.com/r/golang/.rss")
	}

	myApp := app.NewApp(repo)

	fmt.Println("Download articles and saving to database...")
	err = myApp.FetchAll()
	if err != nil {
		log.Fatalf("Error during articles fetch: %v", err)
	}

	fmt.Println("Fetch complete")

	unreadArticles, err := repo.GetUnreadArticles(1, 10)
	if err != nil {
		log.Fatalf("Error reading unread articles: %v", err)
	}

	fmt.Printf("\n Found %d not read for Hacker News.\n\n", len(unreadArticles))
	fmt.Println("--------------------------------------------------")

	for i, article := range unreadArticles {
		fmt.Printf("%d. %s\n", i+1, article.Title)
		fmt.Printf("    %s\n", article.Link)
		fmt.Printf("    Is read: %v | Date: %s\n\n", article.IsRead, article.PublishedAt.Format("2006-01-02 15:04"))
	}
}
