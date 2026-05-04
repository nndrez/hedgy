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

	HelpBarLeft      *tview.TextView
	HelpBarRight     *tview.TextView
	HelpBarContainer *tview.Flex

	ZenMode       bool
	LeftColumn    *tview.Flex
	CenterSection *tview.Flex

	IsFetching   bool
	SpinnerFrame int
}

func NewTUI(backend *app.App, repo storage.Repository) *TUI {
	tview.Borders.TopLeft = '╭'
	tview.Borders.TopRight = '╮'
	tview.Borders.BottomLeft = '╰'
	tview.Borders.BottomRight = '╯'

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

		HelpBarLeft:  tview.NewTextView().SetDynamicColors(true),
		HelpBarRight: tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight),
		ZenMode:      false,
	}

	t.setupUI()
	return t
}

func (t *TUI) Start() error {
	t.loadFeeds()

	go t.startBackgroundWorkers()

	go func() {
		t.IsFetching = true
		err := t.Backend.FetchAll()

		t.App.QueueUpdateDraw(func() {
			t.IsFetching = false
			t.loadFeeds()
			if err != nil {
				t.HelpBarLeft.SetText(fmt.Sprintf(" [red]Fetch error: %v[::-] ", err))
				go func() {
					time.Sleep(4 * time.Second)
					t.App.QueueUpdateDraw(func() { t.updateHelpBar() })
				}()
			}

		})
	}()

	return t.App.Run()
}

func (t *TUI) startBackgroundWorkers() {
	ticker := time.NewTicker(100 * time.Millisecond)
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	for range ticker.C {
		t.App.QueueUpdateDraw(func() {
			currentTime := time.Now().Format("15:04:05")
			status := ""

			if t.IsFetching {
				t.SpinnerFrame = (t.SpinnerFrame + 1) % len(frames)
				status = fmt.Sprintf("[yellow]%s Syncing...[::-] ", frames[t.SpinnerFrame])
			}

			t.HelpBarRight.SetText(fmt.Sprintf("%s[white]%s ", status, currentTime))
		})
	}
}
