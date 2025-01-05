package main

import (
	"fmt"
	"os"
	"sd/pkg/instance"
	"sd/pkg/pages"
	"sd/pkg/profiles"

	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
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
	selectedPosition     string // Track position without committing
	showInstanceSelector bool
	showDeviceSelector   bool
	showProfileSelector  bool
	showPageSelector     bool
	buttonEditor         ButtonEditor
	showButtonEditor     bool
	width                int
	height               int
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
	// Get initial terminal size
	w, h, _ := term.GetSize(0)

	m := model{
		currentInstance:      instance.GetOrCreateInstanceUUID(),
		currentDevice:        "None",
		currentProfile:       "None",
		currentPage:          "None",
		currentButton:        "None",
		selectedPosition:     "1", // Start at first button
		showInstanceSelector: false,
		showDeviceSelector:   false,
		showProfileSelector:  false,
		showPageSelector:     false,
		instanceSelector:     NewInstanceSelector(),
		deviceSelector:       NewDeviceSelector(),
		buttonEditor: NewButtonEditor(
			instance.GetOrCreateInstanceUUID(),
			"None",
			"None",
			"None",
		),
		showButtonEditor: false,
		width:            w,
		height:           h,
	}

	// Set initial editor width
	m.buttonEditor.width = w
	m.buttonEditor.textarea.SetWidth((w - 10) / 2)

	// Initialize profileSelector after m is created
	m.profileSelector = NewProfileSelector(m.currentInstance, m.currentDevice)
	m.pageSelector = NewPageSelector(m.currentInstance, m.currentDevice, m.currentProfile)

	// Set initial widths
	m.instanceSelector.width = w
	m.deviceSelector.width = w
	m.profileSelector.width = w
	m.pageSelector.width = w

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

	// Handle the update for the device selector
	if m.showDeviceSelector {
		// Update the device selector
		cmd = m.deviceSelector.Update(msg)

		// If a device is selected, update the model state
		if device, ok := msg.(DeviceSelected); ok {
			m.currentDevice = string(device)
			m.selectedPosition = "1"   // Reset selected button to 1
			m.showButtonEditor = false // Close button editor
			m.buttonEditor.showEditor = false

			// Recreate the profile selector with the new device
			m.profileSelector = NewProfileSelector(m.currentInstance, m.currentDevice)

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
		cmd = m.profileSelector.Update(msg)

		// Only handle ProfileSelected messages
		if profile, ok := msg.(ProfileSelected); ok {
			m.showProfileSelector = false
			if profile != "" && m.profileSelector.hasSelected {
				m.currentProfile = string(profile)
				m.pageSelector = NewPageSelector(m.currentInstance, m.currentDevice, m.currentProfile)
				return m, nil // Return immediately after handling profile selection
			}
		}
	}

	// Handle the update for the page selector
	if m.showPageSelector {
		cmd = m.pageSelector.Update(msg)
		if page, ok := msg.(PageSelected); ok {
			m.showPageSelector = false
			if page != "" && m.pageSelector.hasSelected {
				m.currentPage = string(page)
				return m, nil // Return immediately after handling page selection
			}
		}
	}

	// Handle button editor updates
	if m.showButtonEditor {
		var editorCmd tea.Cmd
		editorModel, cmd := m.buttonEditor.Update(msg)
		if editor, ok := editorModel.(*ButtonEditor); ok {
			m.buttonEditor = *editor
			editorCmd = cmd
		}
		if cmd != nil {
			return m, editorCmd
		}
	}

	// Handle other messages (e.g., quit, toggling selectors)
	switch msg := msg.(type) {
	case EditorClosing:
		m.showButtonEditor = false
		m.buttonEditor.showEditor = false
		return m, nil
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
			if m.showProfileSelector {
				m.profileSelector.Reset() // Reset the selector state when showing it
			}
			m.showInstanceSelector = false
			m.showDeviceSelector = false
		case "g":
			m.showPageSelector = !m.showPageSelector
			if m.showPageSelector {
				m.pageSelector.Reset() // Reset the selector state when showing it
			}
			m.showInstanceSelector = false
			m.showDeviceSelector = false
			m.showProfileSelector = false
		case "left", "h":
			if m.currentDevice != "None" {
				m.selectedPosition = moveButton(m.selectedPosition, "left", m.currentDevice)
			}
		case "right", "l":
			if m.currentDevice != "None" {
				m.selectedPosition = moveButton(m.selectedPosition, "right", m.currentDevice)
			}
		case "up", "k":
			if m.currentDevice != "None" {
				m.selectedPosition = moveButton(m.selectedPosition, "up", m.currentDevice)
			}
		case "down", "j":
			if m.currentDevice != "None" {
				m.selectedPosition = moveButton(m.selectedPosition, "down", m.currentDevice)
			}
		case "enter":
			if m.currentDevice != "None" && !m.showProfileSelector && !m.showPageSelector {
				m.currentButton = m.selectedPosition
				m.showButtonEditor = true
				m.buttonEditor = NewButtonEditor(
					m.currentInstance,
					m.currentDevice,
					m.currentProfile,
					m.currentPage,
				)
				m.buttonEditor.width = m.width
				m.buttonEditor.textarea.SetWidth((m.width - 10) / 2)
				m.buttonEditor.showEditor = true
				m.buttonEditor.buttonNum = m.currentButton
				m.buttonEditor.LoadButton()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.showButtonEditor {
			m.buttonEditor.width = msg.Width
			m.buttonEditor.textarea.SetWidth((msg.Width - 10) / 2)
		}
		m.instanceSelector.width = msg.Width
		m.deviceSelector.width = msg.Width
		m.profileSelector.width = msg.Width
		m.pageSelector.width = msg.Width
	}

	return m, cmd
}

// View renders the current view based on the state
func (m model) View() string {
	if m.showButtonEditor {
		return m.buttonEditor.View()
	}

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
	deviceView := NewDeviceView(m.currentDevice, m.selectedPosition)

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
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainContent,
		"    ", // Add some spacing
		deviceView.View(),
	)
}

func moveButton(currentButton string, direction string, deviceType string) string {
	// Convert current button from string to int
	current := 1
	if currentButton != "None" {
		if n, err := strconv.Atoi(currentButton); err == nil {
			current = n
		}
	}

	// Get max buttons based on device type
	maxButtons := 32 // XL default
	cols := 8
	if strings.HasPrefix(deviceType, "A0") { // Stream Deck Plus
		maxButtons = 8
		cols = 4
	} else if strings.HasPrefix(deviceType, "FL") { // Stream Deck Pedal
		maxButtons = 3
		cols = 3
	} else if strings.HasPrefix(deviceType, "CL") { // Stream Deck XL
		maxButtons = 32
		cols = 8
	}

	// Calculate new position
	switch direction {
	case "right":
		if current < maxButtons {
			current++
		}
	case "left":
		if current > 1 {
			current--
		}
	case "up":
		if current > cols {
			current -= cols
		}
	case "down":
		if current+cols <= maxButtons {
			current += cols
		}
	}

	return strconv.Itoa(current)
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
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
