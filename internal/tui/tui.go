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
	Pages   *tview.Pages
	Backend *app.App
	Repo    storage.Repository

	FeedList    *tview.List
	ArticleList *tview.List
	ContentView *tview.TextView
	HelpBar     *tview.TextView
}

func NewTUI(backend *app.App, repo storage.Repository) *TUI {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDarkSlateGray
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorLightSkyBlue
	tview.Styles.BorderColor = tcell.ColorDimGray
	tview.Styles.TitleColor = tcell.ColorOrange
	tview.Styles.GraphicsColor = tcell.ColorDimGray

	t := &TUI{
		App:         tview.NewApplication(),
		Pages:       tview.NewPages(),
		Backend:     backend,
		Repo:        repo,
		FeedList:    tview.NewList().ShowSecondaryText(false),
		ArticleList: tview.NewList().ShowSecondaryText(false),
		ContentView: tview.NewTextView().SetDynamicColors(true).SetWordWrap(true),
		HelpBar:     tview.NewTextView().SetDynamicColors(true),
	}

	t.setupUI()
	return t
}

func (t *TUI) setupUI() {
	t.FeedList.SetTitle("Feeds").SetBorder(true)
	t.ArticleList.SetTitle("Articles").SetBorder(true)
	t.ContentView.SetTitle("Content").SetBorder(true)

	t.FeedList.
		SetSelectedBackgroundColor(tcell.ColorDarkCyan).
		SetSelectedTextColor(tcell.ColorWhite).
		SetMainTextColor(tcell.ColorLightGray)

	t.ArticleList.
		SetSelectedBackgroundColor(tcell.ColorDarkMagenta).
		SetSelectedTextColor(tcell.ColorWhite).
		SetMainTextColor(tcell.ColorLightGray)

	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.FeedList, 0, 1, true).
		AddItem(t.ArticleList, 0, 1, false)

	centerSection := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftColumn, 35, 0, true).
		AddItem(t.ContentView, 0, 1, false)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(centerSection, 0, 1, true).
		AddItem(t.HelpBar, 1, 0, false)

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

	t.FeedList.SetFocusFunc(func() {
		t.FeedList.SetBorderColor(tcell.ColorGreen)
		t.ArticleList.SetBorderColor(tcell.ColorDefault)
		t.ContentView.SetBorderColor(tcell.ColorDefault)
		t.HelpBar.SetText(" [::b]Enter[::-]: Open Feed  |  [::b]Tab[::-]: Change Panel  |  [::b]Ctrl+C[::-]: Exit ")
	})

	t.ArticleList.SetFocusFunc(func() {
		t.FeedList.SetBorderColor(tcell.ColorDefault)
		t.ArticleList.SetBorderColor(tcell.ColorGreen)
		t.ContentView.SetBorderColor(tcell.ColorDefault)
		t.HelpBar.SetText(" [::b]Enter[::-]: Read Article  |  [::b]Tab[::-]: Change Panel  |  [::b]Ctrl+C[::-]: Exit ")
	})

	t.ContentView.SetFocusFunc(func() {
		t.FeedList.SetBorderColor(tcell.ColorDefault)
		t.ArticleList.SetBorderColor(tcell.ColorDefault)
		t.ContentView.SetBorderColor(tcell.ColorGreen)
		t.HelpBar.SetText(" [::b]Arrow Keys[::-]: Scroll Text  |  [::b]o[::-]: Open in Browser  |  [::b]Tab[::-]: Change ")
	})

	t.Pages.AddPage("main", mainLayout, true, true)

	t.App.SetRoot(t.Pages, true)
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

func (t *TUI) showAddFeedPrompt() {
	form := tview.NewForm().
		AddInputField("Feed Name:", "", 40, nil, nil).
		AddInputField("URL:", "", 40, nil, nil)

	form.SetBorder(true).
		SetTitle(" Add New Feed ").
		SetTitleAlign(tview.AlignCenter)

	form.AddButton("Save", func() {
		feedName := form.GetFormItemByLabel("Feed Name").(*tview.InputField)
		feedUrl := form.GetFormItemByLabel("URL").(*tview.InputField)

		name := feedName.GetText()
		url := feedUrl.GetText()

		if name == "" || url == "" {
			return
		}

		t.Repo.AddFeed(name, url)
		t.loadFeeds()

		t.Pages.RemovePage("add_feed_modal")
	})

	form.AddButton("Cancel" func() {
		t.Pages.RemovePage("add_feed_modal")
		t.App.SetFocus(t.FeedList)
	})

	modalLayout := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 11, 1, true).
			AddItem(nil, 0, 1, false),
			50, 1, true).
		AddItem(nil, 0, 1, false)

	t.Pages.AddPage("add_feed_modal", modalLayout, true, true)
}
