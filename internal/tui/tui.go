package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
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

func (t *TUI) Start() error {
	t.loadFeeds()

	go func() {
		t.App.QueueUpdateDraw(func() {
			t.HelpBar.SetText(" [yellow]Fetching feeds...[::-] ")
		})

		err := t.Backend.FetchAll()

		t.App.QueueUpdateDraw(func() {
			t.loadFeeds()
			if err != nil {
				t.HelpBar.SetText(fmt.Sprintf(" [red]Fetch error: %v[::-] ", err))
			} else {
				t.HelpBar.SetText(" [green]Feeds updated![::-] ")
				go func() {
					time.Sleep(3 * time.Second)
					t.App.QueueUpdateDraw(func() { t.updateHelpBar() })
				}()
			}
		})
	}()

	return t.App.Run()
}
