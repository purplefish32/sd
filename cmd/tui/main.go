package main

import (
	"fmt"
	"io"
	"os"
	"sd/pkg/instance"
	"sd/pkg/pages"
	"sd/pkg/profiles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// model stores the state of the entire application
type model struct {
	currentInstance    string
	deviceSelector     DeviceSelector
	instanceSelector   InstanceSelector
	currentDevice      string
	currentProfile     string
	currentPage        string
	currentButton      string
	showDevicePicker   bool
	showInstancePicker bool
}

// initialModel initializes the state of the app
func initialModel() model {
	return model{
		currentInstance:    instance.GetOrCreateInstanceUUID(),
		currentDevice:      "None",
		currentProfile:     "None",
		currentPage:        "None",
		currentButton:      "None",
		showDevicePicker:   false,
		showInstancePicker: false,
		deviceSelector:     NewDeviceSelector(),
		instanceSelector:   NewInstanceSelector(),
	}
}

// Init initializes the model (required by tea.Model interface)
func (m model) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates the model state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle the update for the device selector
	if m.showDevicePicker {
		// Update the device selector
		cmd = m.deviceSelector.Update(msg)

		// If a device is selected, update the model state
		if device, ok := msg.(DeviceSelected); ok {
			m.currentDevice = string(device)
			// Process the profile and page updates here if needed
			profile := profiles.GetCurrentProfile(m.currentInstance, m.currentDevice)
			if profile != nil {
				m.currentProfile = profile.ID
				page := pages.GetCurrentPage(m.currentInstance, m.currentDevice, m.currentProfile)
				if page != nil {
					m.currentPage = page.ID
				} else {
					m.currentPage = "Not found"
				}
			} else {
				m.currentProfile = "Not found"
				m.currentPage = "Not found"
			}
			// Close the device picker after selection
			m.showDevicePicker = false
		}
	}

	// Handle the update for the instance selector
	if m.showInstancePicker {
		// Update the instance selector
		cmd = m.instanceSelector.Update(msg)

		// If an instance is selected, update the model state
		if device, ok := msg.(InstanceSelected); ok {
			m.currentInstance = string(device)

			// Close the device picker after selection
			m.showInstancePicker = false
		}
	}

	// Handle other messages (e.g., quit, toggling pickers)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "i":
			m.showInstancePicker = !m.showInstancePicker
			m.showDevicePicker = false
		case "d":
			m.showDevicePicker = !m.showDevicePicker
			m.showInstancePicker = false
		}
	}

	return m, cmd
}

// View renders the current view based on the state
func (m model) View() string {
	if m.showInstancePicker {
		// Show the overlay for the instance picker
		return overlayStyle.Render("\n" + m.instanceSelector.View())
	}

	if m.showDevicePicker {
		// Show the overlay for the device picker
		return overlayStyle.Render("\n" + m.deviceSelector.View())
	}

	// Main view content
	return fmt.Sprintf(`
Current Instance: %s
Current Device: %s
Current Profile: %s
Current Page: %s
Current Button: %s

[i] to change the instance
[d] to change the device
[q] to quit
`, m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.currentButton)
}

// Styling for the overlay
var overlayStyle = lipgloss.NewStyle().Padding(2).Align(lipgloss.Center).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63"))

func main() {
	log.Logger = zerolog.New(io.Discard)

	// Create the initial model and run the program
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
