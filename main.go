package main

import (
	"fmt"
	"log"

	"github.com/nndrez/hedgy/internal/app"
	"github.com/nndrez/hedgy/internal/storage"
)

func main() {
	repo, err := storage.NewJSONStorage()
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

	fmt.Println("Download articles...")
	entries, err := myApp.FetchAll()
	if err != nil {
		log.Fatalf("Error during articles fetch: %v", err)
	}

	fmt.Printf("\n Upload %d articles.\n\n", len(entries))

	fmt.Println("First 10 articles:")
	fmt.Println("--------------------------------------------------")

	limit := min(10, len(entries))

	for i := 0; i < limit; i++ {
		fmt.Printf("%d. %s\n", i+1, entries[i].Title)
		fmt.Printf("    %s\n\n", entries[i].Link)
	}
}
