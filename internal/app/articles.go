package app

import "github.com/nndrez/hedgy/internal/storage"

func (a *App) GetArticles(feedID, limit int, unreadOnly bool) ([]storage.Article, error) {
	return a.repo.GetArticles(feedID, limit, unreadOnly)
}

func (a *App) ToggleReadStatus(articleID int) error {
	return a.repo.ToggleReadStatus(articleID)
}
