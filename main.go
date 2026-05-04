package main

import (
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

	myApp := app.NewApp(repo)
	myTUI := tui.NewTUI(myApp, repo)

	if err := myTUI.Start(); err != nil {
		log.Fatalf("Fatal error in TUI: %v", err)
	}
}
