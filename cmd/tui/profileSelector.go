package main

import (
	"fmt"
	"io"
	"sd/pkg/profiles"
	"sd/pkg/store"

	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

// ProfileSelector represents the state of the profiles selector overlay
type ProfileSelector struct {
	list             list.Model
	selectedItem     string
	instanceID       string
	deviceID         string
	width            int
	hasSelected      bool
	creating         bool
	inputField       textarea.Model
	confirmingDelete bool
	profileToDelete  string
	height           int
}

// NewProfileSelector creates a new instance of ProfileSelector
func NewProfileSelector(instanceID string, deviceID string) ProfileSelector {
	ta := textarea.New()
	ta.Placeholder = "Enter profile name..."
	ta.ShowLineNumbers = false
	ta.SetWidth(30)

	selector := ProfileSelector{
		instanceID:  instanceID,
		deviceID:    deviceID,
		hasSelected: false,
		inputField:  ta,
	}

	// Get terminal width for initial setup
	w, h, _ := term.GetSize(0)
	selector.width = w
	selector.height = h

	// Get profiles using the correct method
	profiles := selector.FetchProfiles()

	profileList := list.New(profiles, profileDelegate{}, w-8, 14)
	profileList.Title = TitleStyle.Render("Select a Profile")
	profileList.SetShowStatusBar(false)
	profileList.SetFilteringEnabled(false)
	profileList.Styles.Title = TitleStyle

	// Add help text for delete command
	profileList.KeyMap.ShowFullHelp.SetEnabled(false)            // Disable full help by default
	profileList.AdditionalShortHelpKeys = func() []key.Binding { // Use ShortHelp instead of FullHelp
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "delete"),
			),
		}
	}

	// Show help
	profileList.SetShowHelp(true)
	profileList.Styles.HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Set initial list width
	profileList.SetSize(w-8, 14)

	// Add "New Profile" as the first item
	newProfileItem := ProfileItem{
		id:    "new",
		title: "+ New Profile",
	}
	allItems := append([]list.Item{newProfileItem}, profiles...)
	profileList.SetItems(allItems)

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

// Add a new message type for profile creation
type CreateProfileMsg string

// Update processes messages and updates the state of the profile selector
func (s *ProfileSelector) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if s.confirmingDelete {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				if err := store.DeleteProfile(s.instanceID, s.deviceID, s.profileToDelete); err != nil {
					log.Error().Err(err).Msg("Failed to delete profile")
				}
				s.confirmingDelete = false
				// Refresh the list
				s.list.SetItems(s.FetchProfiles())
				return nil
			case "n", "N", "esc":
				s.confirmingDelete = false
				return nil
			}
		}
		return nil
	}

	if s.creating {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(s.inputField.Value())
				if name != "" {
					// Create the profile
					profile, err := profiles.CreateProfile(s.instanceID, s.deviceID, name)
					if err != nil {
						log.Error().Err(err).Msg("Failed to create profile")
						s.creating = false
						return nil
					}
					s.creating = false
					s.hasSelected = true
					s.selectedItem = profile.ID
					return func() tea.Msg {
						return ProfileSelected(profile.ID)
					}
				}
			case "esc":
				s.creating = false
				return nil
			}
			var cmd tea.Cmd
			s.inputField, cmd = s.inputField.Update(msg)
			return cmd
		}
		return nil
	}

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
				if selected.id == "new" {
					s.creating = true
					s.inputField.Focus()
					return nil
				}
				s.selectedItem = selected.id
				s.hasSelected = true
				return func() tea.Msg {
					return ProfileSelected(s.selectedItem)
				}
			}
		case "d":
			if selected, ok := s.list.SelectedItem().(ProfileItem); ok && selected.id != "new" {
				s.confirmingDelete = true
				s.profileToDelete = selected.id
				return nil
			}
		}
	}

	return cmd
}

// View renders the profile selector as a string, including the list of profiles
func (s ProfileSelector) View() string {
	if s.confirmingDelete {
		// Create a centered style that uses the full width
		style := ConfirmStyle.Copy().Width(40).Align(lipgloss.Center)

		// Create the confirmation message with vertical centering
		return lipgloss.Place(
			s.width,
			s.height,
			lipgloss.Center,
			lipgloss.Center,
			style.Render(fmt.Sprintf(
				"Are you sure you want to delete this profile? (y/n)\n\n"+
					"This action cannot be undone.",
			)),
		)
	}

	if s.creating {
		style := ListStyle.Copy().Width(40)
		return style.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				"Enter Profile Name:",
				"",
				s.inputField.View(),
				"",
				"Press Enter to create, Esc to cancel",
			),
		)
	}

	style := ListStyle.Copy().Width(s.width - 4)

	s.list.Styles.Title = TitleStyle
	s.list.SetSize(s.width-8, 14)

	return style.Render("Select Profile: (press 'd' to delete)\n\n" + s.list.View())
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

// Add this near the top with other styles or create a new style constant
var ConfirmStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("196")). // Red color
	Padding(1, 2)
