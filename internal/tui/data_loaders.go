package tui

import (
	"fmt"

	"github.com/jaytaylor/html2text"
	"github.com/nndrez/hedgy/internal/storage"
)

func (t *TUI) loadFeeds() {
	t.FeedList.Clear()
	feeds, err := t.Backend.GetFeeds()
	if err != nil {
		t.FeedList.AddItem("Loading feeds error", err.Error(), 'e', nil)
		return
	}

	for _, f := range feeds {
		feed := f
		t.FeedList.AddItem(feed.Name, feed.URL, 0, func() {
			t.loadArticles(feed.ID)
		})
	}

	if len(feeds) > 0 {
		t.loadArticles(feeds[0].ID)
		t.App.SetFocus(t.FeedList)
	}
}

func (t *TUI) loadArticles(feedID int) {
	t.CurrentFeedID = feedID
	t.ArticleList.Clear()
	t.updateArticleTabDisplay()

	articles, err := t.Backend.GetArticles(feedID, 10, t.ShowUnreadOnly)
	if err != nil {
		t.ArticleList.AddItem("Loading error", err.Error(), 'e', nil)
		return
	}

	t.CurrentArticles = articles

	if len(articles) == 0 {
		t.ContentView.SetText("No unread articles in this feed.")
		return
	}

	for _, a := range articles {
		article := a

		displayTitle := article.Title
		if !article.IsRead {
			displayTitle = fmt.Sprintf("[::b]%s[::-]", article.Title)
		}
		t.ArticleList.AddItem(displayTitle, article.PublishedAt.Format(DateFormat), 0, func() {
			t.loadContent(article)
			t.App.SetFocus(t.ContentView)
		})
	}

	t.App.SetFocus(t.ArticleList)
}

func (t *TUI) loadContent(article storage.Article) {
	t.CurrentURL = article.Link

	rawContent := coalesce(article.Content, article.Description, "[Empty feed, follow link to read]")

	plainText, err := html2text.FromString(rawContent, html2text.Options{
		PrettyTables: true,
		OmitLinks:    false,
	})

	if err != nil {
		plainText = rawContent
	}

	header := fmt.Sprintf("[::b]%s[::-]\n\n[blue]%s[-]\n%s\n\n[gray]%s[-]\n\n",
		article.Title,
		article.Link,
		article.PublishedAt.Format(DateFormat),
		"--------------------------------------------------",
	)

	finalText := header + plainText

	t.ContentView.SetText(finalText)
	t.ContentView.ScrollToBeginning()
}

func coalesce(strings ...string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}
	return ""
}
