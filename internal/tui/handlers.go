package tui

import (
	"github.com/gdamore/tcell/v2"
)

func (t *TUI) setupHandlers() {
	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		frontPage, _ := t.Pages.GetFrontPage()
		if frontPage == "add_feed_modal" {
			return event
		}

		switch event.Key() {
		case tcell.KeyTAB:
			if !t.ZenMode {
				t.cycleFocus()
			}
			return nil
		case tcell.KeyCtrlC:
			t.App.Stop()
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'q' {
				t.App.Stop()
				return nil
			}
			if event.Rune() == 'f' {
				t.toggleZenMode()
				return nil
			}
		}

		return event
	})

	t.FeedList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'a' {
			t.showAddFeedPrompt()
			return nil
		}
		return event
	})

	t.ContentView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.App.SetFocus(t.ArticleList)
			return nil
		}
		return event
	})

	t.FeedList.SetFocusFunc(func() {
		t.FeedList.SetBorderColor(tcell.ColorGreen)
		t.ArticleList.SetBorderColor(tcell.ColorDefault)
		t.ContentView.SetBorderColor(tcell.ColorDefault)
		t.updateHelpBar()
	})

	t.ArticleList.SetFocusFunc(func() {
		t.FeedList.SetBorderColor(tcell.ColorDefault)
		t.ArticleList.SetBorderColor(tcell.ColorGreen)
		t.ContentView.SetBorderColor(tcell.ColorDefault)
		t.updateHelpBar()
	})

	t.ContentView.SetFocusFunc(func() {
		t.FeedList.SetBorderColor(tcell.ColorDefault)
		t.ArticleList.SetBorderColor(tcell.ColorDefault)
		t.ContentView.SetBorderColor(tcell.ColorGreen)
		t.updateHelpBar()
	})
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

func (t *TUI) updateHelpBar() {
	if t.ZenMode {
		t.HelpBarLeft.SetText(" [::b]f[::-]: Exit from Zen | [::b]q[::-]: Exit ")
		return
	}

	if t.FeedList.HasFocus() {
		t.HelpBarLeft.SetText(" [::b]Enter[::-]: Open | [::b]a[::-]: Add | [::b]f[::-]: Zen | [::b]Tab[::-]: Panel ")
	} else if t.ArticleList.HasFocus() {
		t.HelpBarLeft.SetText(" [::b]Enter[::-]: Read | [::b]f[::-]: Zen | [::b]Tab[::-]: Panel ")
	} else {
		t.HelpBarLeft.SetText(" [::b]o[::-]: Browser | [::b]f[::-]: Zen | [::b]Tab[::-]: Panel ")
	}
}
