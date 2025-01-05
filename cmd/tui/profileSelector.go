package main

import (
	"fmt"
	"io"
	"sd/pkg/profiles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

// ProfileSelector represents the state of the profiles selector overlay
type ProfileSelector struct {
	list         list.Model
	selectedItem string
	instanceID   string
	deviceID     string
	width        int
	hasSelected  bool
}

// NewProfileSelector creates a new instance of ProfileSelector
func NewProfileSelector(instanceID string, deviceID string) ProfileSelector {
	selector := ProfileSelector{
		instanceID:  instanceID,
		deviceID:    deviceID,
		hasSelected: false,
	}

	// Get terminal width for initial setup
	w, _, _ := term.GetSize(0)
	selector.width = w

	// Get profiles using the correct method
	profiles := selector.FetchProfiles()

	profileList := list.New(profiles, profileDelegate{}, w-8, 14)
	profileList.Title = TitleStyle.Render("Select a Profile")
	profileList.SetShowStatusBar(false)
	profileList.SetFilteringEnabled(false)
	profileList.Styles.Title = TitleStyle

	// Set initial list width
	profileList.SetSize(w-8, 14)

	selector.list = profileList
	return selector
}

// Init is required by the tea.Model interface, but not used in this case
func (s ProfileSelector) Init() tea.Cmd {
	return nil
}

func (s ProfileSelector) SelectedProfile() string {
	return s.selectedItem
}

// Define a custom message type for when a profile is selected
type ProfileSelected string

// Update processes messages and updates the state of the profile selector
func (s *ProfileSelector) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// First handle list updates
	s.list, cmd = s.list.Update(msg)

	// Then handle our custom logic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.hasSelected = false
			s.selectedItem = ""
			return func() tea.Msg {
				return ProfileSelected("")
			}
		case "enter":
			if selected, ok := s.list.SelectedItem().(ProfileItem); ok {
				s.selectedItem = selected.id
				s.hasSelected = true
				return func() tea.Msg {
					return ProfileSelected(s.selectedItem)
				}
			}
		}
	}

	return cmd
}

// View renders the profile selector as a string, including the list of profiles
func (s ProfileSelector) View() string {
	style := ListStyle.Copy().Width(s.width - 4)

	s.list.Styles.Title = TitleStyle
	s.list.SetSize(s.width-8, 14)

	return style.Render("Select Profile:\n\n" + s.list.View())
}

// profileDelegate handles rendering of each item in the profile selection list
type profileDelegate struct{}

func (d profileDelegate) Height() int                             { return 1 }
func (d profileDelegate) Spacing() int                            { return 0 }
func (d profileDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d profileDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ProfileItem)
	if !ok {
		return
	}

	str := i.title
	if index == m.Index() {
		str = "> " + str
	}

	fmt.Fprint(w, str)
}

func (s *ProfileSelector) FetchProfiles() []list.Item {
	profiles, err := profiles.GetProfiles(s.instanceID, s.deviceID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch profiles")
		return []list.Item{}
	}

	items := make([]list.Item, len(profiles))
	for i, profile := range profiles {
		items[i] = ProfileItem{
			id:    profile.ID,
			title: profile.Name,
		}
	}
	return items
}

// Add this struct definition near the top of the file with other types
type ProfileItem struct {
	id    string
	title string
}

// Add the FilterValue method for the ProfileItem
func (i ProfileItem) FilterValue() string {
	return i.title
}

// Add this method to ProfileSelector
func (s *ProfileSelector) Reset() {
	s.hasSelected = false
	s.selectedItem = ""
}
