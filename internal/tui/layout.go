package tui

import (
	"fmt"

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

	t.ArticleTabs = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	articleSection := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.ArticleTabs, 1, 0, false).
		AddItem(t.ArticleList, 0, 1, true)

	t.LeftColumn = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.FeedList, 0, 1, true).
		AddItem(articleSection, 0, 1, false)

	t.CenterSection = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.LeftColumn, 35, 0, true).
		AddItem(t.ContentView, 0, 1, false)

	t.HelpBarContainer = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.HelpBarLeft, 0, 1, false).
		AddItem(t.HelpBarRight, 25, 0, false)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.CenterSection, 0, 1, true).
		AddItem(t.HelpBarContainer, 1, 0, false)

	t.setupHandlers()

	t.Pages.AddPage("main", mainLayout, true, true)
	t.App.SetRoot(t.Pages, true)
}

func (t *TUI) toggleZenMode() {
	t.ZenMode = !t.ZenMode

	t.CenterSection.Clear()

	if t.ZenMode {
		t.CenterSection.AddItem(t.ContentView, 0, 1, true)
		t.App.SetFocus(t.ContentView)
	} else {
		t.CenterSection.AddItem(t.LeftColumn, 35, 0, true).
			AddItem(t.ContentView, 0, 1, false)
		t.App.SetFocus(t.ArticleList)
	}

	t.updateHelpBar()
}

func (t *TUI) updateArticleTabDisplay() {
	var allTab, unreadTab string

	if t.ShowUnreadOnly {
		allTab = "[white][ All ][-]"
		unreadTab = "[yellow::b][ Unread ][-]"
	} else {
		allTab = "[yellow::b][ All ][-]"
		unreadTab = "[white][ Unread ][-]"
	}

	t.ArticleTabs.SetText(fmt.Sprintf("%s   %s", allTab, unreadTab))
}
