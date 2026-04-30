package main

import (
	"fmt"
	"log"

	"github.com/nndrez/hedgy/internal/app"
	"github.com/nndrez/hedgy/internal/storage"
	"github.com/nndrez/hedgy/internal/tui"
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
		repo.AddFeed("antirez", "https://antirez.com/rss")
	}

	myApp := app.NewApp(repo)
	_ = myApp.FetchAll()

	myTUI := tui.NewTUI(myApp, repo)
	if err := myTUI.Start(); err != nil {
		log.Fatalf("Fatal error in TUI: %v", err)
	}
}
