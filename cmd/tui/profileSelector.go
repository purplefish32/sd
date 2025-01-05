package main

import (
	"fmt"
	"io"
	"sd/pkg/profiles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

// Item represents a single item in the profile list
type Profile string

// FilterValue returns the value used for filtering
func (p Profile) FilterValue() string {
	return string(p) // Return the string value of the Item
}

// ProfileSelector represents the state of the profiles selector overlay
type ProfileSelector struct {
	list         list.Model
	selectedItem string
	instanceID   string
	deviceID     string
}

// Profiles list (note: using Item type here)
func FetchProfiles(instanceID string, deviceID string) []list.Item {

	log.Debug().Msg("FetchProfiles CALLED")
	log.Debug().Msg(instanceID)
	log.Debug().Msg(deviceID)

	profiles.GetProfiles(instanceID, deviceID)

	// Example dynamic profile fetch (replace with your actual logic)
	return []list.Item{
		Profile("Profile 1"),
		Profile("Profile 2"),
		Profile("Profile 3"),
	}
}

// NewProfileSelector creates a new instance of ProfileSelector
func NewProfileSelector(instanceID string, deviceID string) ProfileSelector {
	selector := ProfileSelector{
		instanceID: instanceID,
		deviceID:   deviceID,
	}

	profiles := selector.FetchProfiles()

	profileList := list.New(profiles, profileDelegate{}, 20, 14)
	profileList.Title = TitleStyle.Render("Select a Profile")
	profileList.SetShowStatusBar(false)
	profileList.SetFilteringEnabled(false)
	profileList.Styles.Title = TitleStyle

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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return func() tea.Msg {
				return ProfileSelected("")
			}
		case "enter":
			// When a profile is selected, store it
			selected, ok := s.list.SelectedItem().(ProfileItem)
			if ok {
				s.selectedItem = selected.id
			}
			// Return a command that signals the parent model to handle the selected profile
			return func() tea.Msg {
				return ProfileSelected(s.selectedItem)
			}
		}
	}

	// Update the profile list whenever the state changes
	profiles := s.FetchProfiles()
	s.list.SetItems(profiles)

	s.list, cmd = s.list.Update(msg)
	return cmd
}

// View renders the profile selector as a string, including the list of profiles
func (s ProfileSelector) View() string {
	return ListStyle.Render(s.list.View())
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
