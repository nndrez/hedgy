package app

import "github.com/nndrez/hedgy/internal/storage"

func (a *App) GetFeeds() ([]storage.Feed, error) {
	return a.repo.GetFeeds()
}

func (a *App) AddFeed(name, url string) error {
	return a.repo.AddFeed(name, url)
}

func (a *App) DeleteFeed(id int) error {
	return a.repo.DeleteFeed(id)
}
