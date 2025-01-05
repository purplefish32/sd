package main

import (
	"fmt"
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
	currentInstance      string
	deviceSelector       DeviceSelector
	instanceSelector     InstanceSelector
	profileSelector      ProfileSelector
	pageSelector         PageSelector
	currentDevice        string
	currentProfile       string
	currentPage          string
	currentButton        string
	showInstanceSelector bool
	showDeviceSelector   bool
	showProfileSelector  bool
	showPageSelector     bool
}

// Getter for currentInstance
func (m *model) GetCurrentInstance() string {
	return m.currentInstance
}

// Getter for currentDevice
func (m *model) GetCurrentDevice() string {
	return m.currentDevice
}

// initialModel initializes the state of the app
func initialModel() model {
	m := model{
		currentInstance:      instance.GetOrCreateInstanceUUID(),
		currentDevice:        "None",
		currentProfile:       "None",
		currentPage:          "None",
		currentButton:        "None",
		showInstanceSelector: false,
		showDeviceSelector:   false,
		showProfileSelector:  false,
		showPageSelector:     false,
		instanceSelector:     NewInstanceSelector(),
		deviceSelector:       NewDeviceSelector(),
	}
	// Initialize profileSelector after m is created
	m.profileSelector = NewProfileSelector(m.currentInstance, m.currentDevice)
	m.pageSelector = NewPageSelector(m.currentInstance, m.currentDevice, m.currentProfile)
	return m
}

// Init initializes the model (required by tea.Model interface)
func (m model) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates the model state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	log.Debug().Msg("UPDATE")

	// This recreation on every Update call could be optimized
	m.profileSelector = NewProfileSelector(m.currentInstance, m.currentDevice) // TODO

	// Handle the update for the device selector
	if m.showDeviceSelector {
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
			// Close the device selector after selection
			m.showDeviceSelector = false
		}
	}

	// Handle the update for the instance selector
	if m.showInstanceSelector {
		// Update the instance selector
		cmd = m.instanceSelector.Update(msg)

		// If an instance is selected, update the model state
		if device, ok := msg.(InstanceSelected); ok {
			m.currentInstance = string(device)

			// Close the instance selector after selection
			m.showInstanceSelector = false
		}
	}

	// Handle the update for the profile selector
	if m.showProfileSelector {
		// Update the profile selector
		cmd = m.profileSelector.Update(msg)

		// If a profile is selected, update the model state
		if device, ok := msg.(ProfileSelected); ok {
			m.currentProfile = string(device)
			// Recreate the page selector with the new profile
			m.pageSelector = NewPageSelector(m.currentInstance, m.currentDevice, m.currentProfile)

			// Close the profile selector after selection
			m.showProfileSelector = false
		}
	}

	// Handle the update for the page selector
	if m.showPageSelector {
		cmd = m.pageSelector.Update(msg)
		if page, ok := msg.(PageSelected); ok {
			m.currentPage = string(page)
			m.showPageSelector = false
		}
	}

	// Handle other messages (e.g., quit, toggling selectors)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "i":
			m.showInstanceSelector = !m.showInstanceSelector
			m.showDeviceSelector = false
			m.showProfileSelector = false
		case "d":
			m.showDeviceSelector = !m.showDeviceSelector
			m.showInstanceSelector = false
			m.showProfileSelector = false
		case "p":
			m.showProfileSelector = !m.showProfileSelector
			m.showInstanceSelector = false
			m.showDeviceSelector = false
		case "g":
			m.showPageSelector = !m.showPageSelector
			m.showInstanceSelector = false
			m.showDeviceSelector = false
			m.showProfileSelector = false
		}
	}

	return m, cmd
}

// View renders the current view based on the state
func (m model) View() string {
	if m.showInstanceSelector {
		return m.instanceSelector.View()
	}

	if m.showDeviceSelector {
		return m.deviceSelector.View()
	}

	if m.showProfileSelector {
		return m.profileSelector.View()
	}

	if m.showPageSelector {
		return m.pageSelector.View()
	}

	// Create device view
	deviceView := NewDeviceView(m.currentDevice)

	// Main content with device view
	mainContent := fmt.Sprintf(`
Current Instance: %s
Current Device: %s
Current Profile: %s
Current Page: %s
Current Button: %s

[i] to change the instance
[d] to change the device
[p] to change the profile
[g] to change the page
[q] to quit
`, m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.currentButton)

	// If we have a device selected, show the device view next to the main content
	if m.currentDevice != "None" {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			mainContent,
			"    ", // Add some spacing
			deviceView.View(),
		)
	}

	return mainContent
}

func main() {
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Error opening log file:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Set the global logger to log to the file
	log.Logger = zerolog.New(logFile).With().Timestamp().Logger()

	// Create the initial model and run the program
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
