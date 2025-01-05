package main

import (
	"fmt"
	"io"
	"sd/pkg/pages"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

// PageSelector represents the state of the page selector overlay
type PageSelector struct {
	list         list.Model
	selectedItem string
	instanceID   string
	deviceID     string
	profileID    string
}

// PageItem represents a single page in the list
type PageItem struct {
	id    string
	title string
}

// FilterValue returns the value used for filtering
func (i PageItem) FilterValue() string {
	return i.title
}

// Define a custom message type for when a page is selected
type PageSelected string

// NewPageSelector creates a new instance of PageSelector
func NewPageSelector(instanceID, deviceID, profileID string) PageSelector {
	selector := PageSelector{
		instanceID: instanceID,
		deviceID:   deviceID,
		profileID:  profileID,
	}

	pages := selector.FetchPages()

	pageList := list.New(pages, pageDelegate{}, 20, 14)
	pageList.Title = TitleStyle.Render("Select a Page")
	pageList.SetShowStatusBar(false)
	pageList.SetFilteringEnabled(false)
	pageList.Styles.Title = TitleStyle

	selector.list = pageList
	return selector
}

func (s *PageSelector) FetchPages() []list.Item {
	log.Debug().
		Str("instanceID", s.instanceID).
		Str("deviceID", s.deviceID).
		Str("profileID", s.profileID).
		Msg("Fetching pages")

	pages, err := pages.GetPages(s.instanceID, s.deviceID, s.profileID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch pages")
		return []list.Item{}
	}

	log.Debug().
		Int("pageCount", len(pages)).
		Interface("pages", pages).
		Msg("Retrieved pages")

	items := make([]list.Item, len(pages))
	for i, page := range pages {
		items[i] = PageItem{
			id:    page.ID,
			title: fmt.Sprintf("Page %d", i+1),
		}
	}
	return items
}

// Init is required by the tea.Model interface
func (s PageSelector) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates the state of the page selector
func (s *PageSelector) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return func() tea.Msg {
				return PageSelected("")
			}
		case "enter":
			selected, ok := s.list.SelectedItem().(PageItem)
			if ok {
				s.selectedItem = selected.id
			}
			return func() tea.Msg {
				return PageSelected(s.selectedItem)
			}
		}
	}

	// Update the page list whenever the state changes
	pages := s.FetchPages()
	s.list.SetItems(pages)

	s.list, cmd = s.list.Update(msg)
	return cmd
}

// View renders the page selector
func (s PageSelector) View() string {
	return ListStyle.Render(s.list.View())
}

// pageDelegate handles rendering of each item in the page selection list
type pageDelegate struct{}

func (d pageDelegate) Height() int                             { return 1 }
func (d pageDelegate) Spacing() int                            { return 0 }
func (d pageDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d pageDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(PageItem)
	if !ok {
		return
	}

	str := i.title
	if index == m.Index() {
		str = "> " + str
	}

	fmt.Fprint(w, str)
}
