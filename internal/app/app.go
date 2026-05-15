package app

import "github.com/nndrez/hedgy/internal/storage"

type App struct {
	repo storage.Repository
}

func NewApp(repo storage.Repository) *App {
	return &App{repo: repo}
}
