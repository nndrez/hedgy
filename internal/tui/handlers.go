package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

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
		case tcell.KeyBacktab:
			if !t.ZenMode {
				t.reverseCycleFocus()
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
		}

		return event
	})

	t.FeedList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'a' {
			t.showAddFeedPrompt()
			return nil
		}
		if event.Rune() == 'r' {
			t.refreshAllFeeds()
			return nil
		}
		return event
	})

	t.ArticleList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'r' {
			t.refreshCurrentFeed()
			return nil
		}
		if event.Rune() == 'm' {
			index := t.ArticleList.GetCurrentItem()
			if index >= 0 && index < len(t.CurrentArticles) {
				article := t.CurrentArticles[index]

				err := t.Repo.ToggleReadStatus(article.ID)
				if err != nil {
					t.HelpBarLeft.SetText(fmt.Sprintf(" [red]Update error: %v[::-] ", err))
					return nil
				}

				t.loadArticles(t.CurrentFeedID)
				t.ArticleList.SetCurrentItem(index)
			}
			return nil
		}
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			index := t.ArticleList.GetCurrentItem()
			if index >= 0 && index < len(t.CurrentArticles) {
				article := t.CurrentArticles[index]

				if !article.IsRead {
					err := t.Repo.ToggleReadStatus(article.ID)
					if err != nil {
						t.HelpBarLeft.SetText(fmt.Sprintf(" [red]Update error: %v[::-] ", err))
						return nil
					}
				}

				t.loadArticles(t.CurrentFeedID)
				t.ArticleList.SetCurrentItem(index)
				t.loadContent(article)
				t.App.SetFocus(t.ContentView)
			}
			return nil
		}
		if event.Rune() == 'f' {
			t.ShowUnreadOnly = !t.ShowUnreadOnly
			t.loadArticles(t.CurrentFeedID)
			return nil
		}
		return event
	})

	t.ContentView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if t.ZenMode {
				t.toggleZenMode()
			} else {
				t.App.SetFocus(t.ArticleList)
			}
			return nil
		}

		switch event.Rune() {
		case 'f':
			t.toggleZenMode()
			return nil
		case 'o':
			if t.CurrentURL == "" {
				t.HelpBarLeft.SetText(" [red]Error: No URL found for this article[::-] ")
			} else {
				err := openBrowser(t.CurrentURL)
				if err != nil {
					t.HelpBarLeft.SetText(fmt.Sprintf(" [red]Error launching browser: %v[::-] ", err))
				} else {
					t.HelpBarLeft.SetText(fmt.Sprintf(" [green]Opening browser: %s[::-] ", t.CurrentURL))
				}
			}

			go func() {
				time.Sleep(4 * time.Second)
				t.App.QueueUpdateDraw(func() { t.updateHelpBar() })
			}()
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

// feed -> article -> context
func (t *TUI) cycleFocus() {
	if t.FeedList.HasFocus() {
		t.App.SetFocus(t.ArticleList)
	} else if t.ArticleList.HasFocus() {
		t.App.SetFocus(t.ContentView)
	} else {
		t.App.SetFocus(t.FeedList)
	}
}

// feed -> context -> article
func (t *TUI) reverseCycleFocus() {
	if t.FeedList.HasFocus() {
		t.App.SetFocus(t.ContentView)
	} else if t.ArticleList.HasFocus() {
		t.App.SetFocus(t.FeedList)
	} else {
		t.App.SetFocus(t.ArticleList)
	}
}

func (t *TUI) updateHelpBar() {
	if t.ZenMode {
		t.HelpBarLeft.SetText(" [::b]f[::-]: Exit from Zen | [::b]q[::-]: Exit ")
		return
	}

	if t.FeedList.HasFocus() {
		t.HelpBarLeft.SetText(" [::b]Enter[::-]: Open | [::b]r[::-]: Refresh | [::b]a[::-]: Add | [::b]Tab[::-]: Panel ")
	} else if t.ArticleList.HasFocus() {
		t.HelpBarLeft.SetText(" [::b]Enter[::-]: Read | [::b]Space[::-]: Read | [::b]m[::-]: Mark | [::b]f[::-]: Filter | [::b]r[::-]: Refresh | [::b]Tab[::-]: Panel ")
	} else {
		t.HelpBarLeft.SetText(" [::b]o[::-]: Browser | [::b]f[::-]: Zen | [::b]Tab[::-]: Panel ")
	}
}

func openBrowser(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	isWSL := false
	if b, err := os.ReadFile("/proc/version"); err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft") {
		isWSL = true
	}

	var cmd string
	var args []string

	if isWSL {
		cmd = "powershell.exe"
		args = []string{"-NoProfile", "-Command", fmt.Sprintf("Start-Process '%s'", url)}
	} else {
		switch runtime.GOOS {
		case "windows":
			cmd = "powershell"
			args = []string{"-NoProfile", "-Command", fmt.Sprintf("Start-Process '%s'", url)}
		case "darwin":
			cmd = "open"
			args = []string{url}
		default:
			cmd = "xdg-open"
			args = []string{url}
		}
	}

	return exec.Command(cmd, args...).Start()
}
