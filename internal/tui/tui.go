package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/jaytaylor/html2text"
	"github.com/nndrez/hedgy/internal/app"
	"github.com/nndrez/hedgy/internal/storage"
	"github.com/rivo/tview"
)

type TUI struct {
	App     *tview.Application
	Backend *app.App
	Repo    storage.Repository

	FeedList    *tview.List
	ArticleList *tview.List
	ContentView *tview.TextView
}

func NewTUI(backend *app.App, repo storage.Repository) *TUI {
	t := &TUI{
		App:         tview.NewApplication(),
		Backend:     backend,
		Repo:        repo,
		FeedList:    tview.NewList().ShowSecondaryText(false),
		ArticleList: tview.NewList().ShowSecondaryText(false),
		ContentView: tview.NewTextView().SetDynamicColors(true).SetWordWrap(true),
	}

	t.setupUI()
	return t
}

func (t *TUI) setupUI() {
	t.FeedList.SetTitle("Feeds").SetBorder(true)
	t.ArticleList.SetTitle("Articles").SetBorder(true)
	t.ContentView.SetTitle("Content").SetBorder(true)

	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.FeedList, 0, 1, true).
		AddItem(t.ArticleList, 0, 1, false)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftColumn, 35, 0, true).
		AddItem(t.ContentView, 0, 1, false)

	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			t.cycleFocus()
			return nil
		case tcell.KeyCtrlC:
			t.App.Stop()
			return nil
		}

		return event
	})

	t.App.SetRoot(mainLayout, true)
}

func (t *TUI) cycleFocus() {
	if t.FeedList.HasFocus() {
		t.App.SetFocus(t.ArticleList)
	} else if t.ArticleList.HasFocus() {
		t.App.SetFocus(t.ContentView)
	} else {
		t.App.SetFocus(t.FeedList)
	}
}

func (t *TUI) Start() error {
	t.loadFeeds()
	return t.App.Run()
}

func (t *TUI) loadFeeds() {
	t.FeedList.Clear()
	feeds, err := t.Repo.GetFeeds()
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
}

func (t *TUI) loadArticles(feedID int) {
	t.ArticleList.Clear()
	t.ContentView.Clear()

	articles, err := t.Repo.GetUnreadArticles(feedID, 10)
	if err != nil {
		t.ArticleList.AddItem("Loading error", err.Error(), 'e', nil)
		return
	}

	if len(articles) == 0 {
		t.ContentView.SetText("No unread articles in this feed.")
		return
	}

	for _, a := range articles {
		article := a
		t.ArticleList.AddItem(article.Title, article.PublishedAt.Format("02-01-2006 15:04"), 0, func() {
			t.loadContent(article)
		})
	}

	t.App.SetFocus(t.ArticleList)

	t.loadContent(articles[0])
}

func (t *TUI) loadContent(article storage.Article) {
	rawContent := coalesce(article.Content, article.Description, "[Empty feed, follow link to read]")

	plainText, err := html2text.FromString(rawContent, html2text.Options{
		PrettyTables: true,
		OmitLinks:    false,
	})

	if err != nil {
		plainText = rawContent
	}

	header := fmt.Sprintf("[::b]%s[::-]\n\n🔗 [blue]%s[-]\n📅 %s\n\n[gray]%s[-]\n\n",
		article.Title,
		article.Link,
		article.PublishedAt.Format("02 Jan 2006 15:04"),
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
