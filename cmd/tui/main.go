package main

import (
	"fmt"
	"io"
	"os"
	"sd/pkg/instance"

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

func Model() model {
	return model{
		currentInstance:    instance.GetOrCreateInstanceUUID(),
		deviceSelector:     NewDeviceSelector(),
		instanceSelector:   NewInstanceSelector(),
		currentDevice:      "None",
		currentProfile:     "None",
		currentPage:        "None",
		currentButton:      "None",
		showDevicePicker:   false, // Start with device picker hidden
		showInstancePicker: false, // Start with instance picker hidden

	}
}

// Init initializes the model (required by tea.Model interface)
func (m model) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates the model state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Update instanceSelector and deviceSelector based on the active picker
	if m.showInstancePicker {
		m.instanceSelector, cmd = m.instanceSelector.Update(msg)
	} else if m.showDevicePicker {
		m.deviceSelector, cmd = m.deviceSelector.Update(msg)
	}

	// Handle key messages for selecting or quitting
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "i":
			// Toggle the instance picker visibility
			m.showInstancePicker = !m.showInstancePicker
			m.showDevicePicker = false // Hide device picker when showing instance picker
		case "d":
			// Toggle the device picker visibility
			m.showDevicePicker = !m.showDevicePicker
			m.showInstancePicker = false // Hide instance picker when showing device picker
		case "enter":
			// Select instance and device, hide the overlays
			if m.showInstancePicker {
				i, ok := m.instanceSelector.list.SelectedItem().(Item)
				if ok {
					m.currentInstance = string(i)
				}
				m.showInstancePicker = false
			}

			if m.showDevicePicker {
				d, ok := m.deviceSelector.list.SelectedItem().(Item)
				if ok {
					m.currentDevice = string(d)
				}
				m.showDevicePicker = false
			}
		}
	}

	return m, cmd
}

// View renders the current view based on the state
func (m model) View() string {
	if m.showInstancePicker {
		// Show the overlay for the device picker
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
	m := Model()
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
