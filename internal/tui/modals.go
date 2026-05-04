package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/nndrez/hedgy/internal/storage"
	"github.com/rivo/tview"
)

func (t *TUI) showAddFeedPrompt() {
	form := tview.NewForm().
		AddInputField("Feed Name", "", 0, nil, nil).
		AddInputField("URL", "", 0, nil, nil)

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

		go func() {
			t.IsFetching = true

			feeds, _ := t.Repo.GetFeeds()
			var newFeed storage.Feed
			for _, f := range feeds {
				if f.URL == url {
					newFeed = f
					break
				}
			}

			if newFeed.ID != 0 {
				err := t.Backend.FetchFeed(newFeed)

				t.App.QueueUpdateDraw(func() {
					t.IsFetching = false
					if err != nil {
						t.HelpBarLeft.SetText(fmt.Sprintf(" [red]Fetch error: %v[::-] ", err))
					} else {
						t.HelpBarLeft.SetText(fmt.Sprintf(" [green]Feed added: %s[::-] ", name))
					}

					go func() {
						time.Sleep(4 * time.Second)
						t.App.QueueUpdateDraw(func() { t.updateHelpBar() })
					}()
				})
			} else {
				t.App.QueueUpdateDraw(func() { t.IsFetching = false })
			}
		}()

		t.Pages.RemovePage("add_feed_modal")
		t.App.SetFocus(t.FeedList)
		t.updateHelpBar()
	})

	form.AddButton("Cancel", func() {
		t.Pages.RemovePage("add_feed_modal")
		t.App.SetFocus(t.FeedList)
		t.updateHelpBar()
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.Pages.RemovePage("add_feed_modal")
			t.App.SetFocus(t.FeedList)
			t.updateHelpBar()
			return nil
		}
		return event
	})

	modalLayout := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 11, 1, true).
			AddItem(nil, 0, 1, false),
			0, 2, true).
		AddItem(nil, 0, 1, false)

	t.Pages.AddPage("add_feed_modal", modalLayout, true, true)
}
