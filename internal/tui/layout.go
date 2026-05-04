package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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

	t.setupHandlers()

	t.Pages.AddPage("main", mainLayout, true, true)
	t.App.SetRoot(t.Pages, true)
}
