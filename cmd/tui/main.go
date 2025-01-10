package main

import (
	"fmt"
	"os"
	"sd/pkg/buttons"
	"sd/pkg/instance"
	"sd/pkg/natsconn"
	"sd/pkg/pages"
	"sd/pkg/profiles"
	"sd/pkg/store"

	"strconv"
	"strings"

	"sd/pkg/actions"
	"sd/pkg/plugins/browser"
	"sd/pkg/plugins/command"
	"sd/pkg/plugins/keyboard"

	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

// model stores the state of the entire application
type model struct {
	currentInstance        string
	deviceSelector         DeviceSelector
	instanceSelector       InstanceSelector
	profileSelector        ProfileSelector
	pageSelector           PageSelector
	currentDevice          string
	currentProfile         string
	currentPage            string
	currentButton          string
	selectedPosition       string // Track position without committing
	showInstanceSelector   bool
	showDeviceSelector     bool
	showProfileSelector    bool
	showPageSelector       bool
	buttonEditor           ButtonEditor
	showButtonEditor       bool
	width                  int
	height                 int
	currentProfileName     string
	showDeleteConfirmation bool
	buttonClipboard        string // Store copied button data
	showPasteConfirmation  bool
	swapMode               bool   // True when in swap mode
	swapSourceButton       string // Store the first button position for swap
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
		currentProfileName:   "None",
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
		showButtonEditor:       false,
		width:                  w,
		height:                 h,
		showDeleteConfirmation: false,
		buttonClipboard:        "",
		showPasteConfirmation:  false,
		swapMode:               false,
		swapSourceButton:       "",
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

	// Handle the update for the device selector
	if m.showDeviceSelector {
		// Update the device selector
		cmd = m.deviceSelector.Update(msg)

		// If a device is selected, update the model state
		if device, ok := msg.(DeviceSelected); ok {
			m.currentDevice = string(device)
			m.selectedPosition = "1"
			m.showButtonEditor = false
			m.buttonEditor.showEditor = false

			m.profileSelector = NewProfileSelector(m.currentInstance, m.currentDevice)

			// First get the current profile ID
			if currentProfile := profiles.GetCurrentProfile(m.currentInstance, m.currentDevice); currentProfile != nil {
				// Then get the full profile details using store
				if profile, err := store.GetProfile(m.currentInstance, m.currentDevice, currentProfile.ID); err == nil {
					m.currentProfile = profile.ID
					m.currentProfileName = profile.Name

					page := pages.GetCurrentPage(m.currentInstance, m.currentDevice, m.currentProfile)
					if page != nil {
						m.currentPage = page.ID
					} else {
						m.currentPage = "Not found"
					}
				}
			} else {
				m.currentProfile = "Not found"
				m.currentProfileName = "Not found"
				m.currentPage = "Not found"
			}
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

			// Reset all device-related state
			m.currentDevice = "None"
			m.currentProfile = "None"
			m.currentProfileName = "None"
			m.currentPage = "None"
			m.currentButton = "None"

			// Reset selectors
			m.deviceSelector = NewDeviceSelector()
			m.profileSelector = NewProfileSelector(m.currentInstance, m.currentDevice)
			m.pageSelector = NewPageSelector(m.currentInstance, m.currentDevice, m.currentProfile)

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

				// Get the profile to show its name and handle pages
				if prof, err := store.GetProfile(m.currentInstance, m.currentDevice, m.currentProfile); err == nil {
					m.currentProfileName = prof.Name
					if len(prof.Pages) > 0 {
						m.currentPage = prof.Pages[0].ID
						pages.SetCurrentPage(m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage)
					}
				}

				m.pageSelector = NewPageSelector(m.currentInstance, m.currentDevice, m.currentProfile)
				return m, nil
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
		// Only handle global commands if no popup is open
		if !m.showInstanceSelector && !m.showDeviceSelector &&
			!m.showProfileSelector && !m.showPageSelector && !m.showButtonEditor {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "i":
				m.showInstanceSelector = true
			case "d":
				m.showDeviceSelector = true
			case "p":
				if m.currentDevice != "None" {
					// Create a fresh profile selector with current state
					m.profileSelector = NewProfileSelector(m.currentInstance, m.currentDevice)
					m.showProfileSelector = true
					if m.showProfileSelector {
						m.profileSelector.Reset()
					}
				}
			case "g":
				if m.currentDevice != "None" {
					m.showPageSelector = true
					if m.showPageSelector {
						m.pageSelector.Reset()
					}
				}
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
				if m.swapMode {
					if m.swapSourceButton != m.selectedPosition {
						sourceKey := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
							m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.swapSourceButton)
						targetKey := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
							m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.selectedPosition)

						_, kv := natsconn.GetNATSConn()
						sourceButton, err := buttons.GetButton(sourceKey)
						targetButton, err2 := buttons.GetButton(targetKey)

						// Swap the buttons
						if err == nil {
							if data, err := json.Marshal(sourceButton); err == nil {
								kv.Put(targetKey, data)
								// Delete the old buffer to force an update
								kv.Delete(targetKey + ".buffer")
							}
						} else {
							kv.Delete(targetKey)
							kv.Delete(targetKey + ".buffer")
						}

						if err2 == nil {
							if data, err := json.Marshal(targetButton); err == nil {
								kv.Put(sourceKey, data)
								// Delete the old buffer to force an update
								kv.Delete(sourceKey + ".buffer")
							}
						} else {
							kv.Delete(sourceKey)
							kv.Delete(sourceKey + ".buffer")
						}

						log.Debug().
							Str("source", m.swapSourceButton).
							Str("target", m.selectedPosition).
							Msg("Swapped buttons")
					}
					m.swapMode = false
					m.swapSourceButton = ""
					return m, nil
				} else if m.currentDevice != "None" {
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
			case "x": // Delete
				if m.currentDevice != "None" {
					key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
						m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.selectedPosition)

					// Check if button exists before showing confirmation
					if _, err := buttons.GetButton(key); err == nil {
						m.showDeleteConfirmation = true
					}
				}
			case "c": // Copy
				if m.currentDevice != "None" {
					key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
						m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.selectedPosition)

					// If button doesn't exist, store empty string to represent a blank button
					if _, err := buttons.GetButton(key); err != nil {
						m.buttonClipboard = "{}"
						log.Debug().Str("copied_button", m.selectedPosition).Msg("Copied blank button")
					} else {
						if button, err := buttons.GetButton(key); err == nil {
							if data, err := json.Marshal(button); err == nil {
								m.buttonClipboard = string(data)
								log.Debug().Str("copied_button", m.selectedPosition).Msg("Button copied")
							}
						}
					}
				}
			case "v": // Paste
				if m.currentDevice != "None" && m.buttonClipboard != "" {
					key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
						m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.selectedPosition)

					// Check if we're pasting a blank button
					if m.buttonClipboard == "{}" {
						// Only show confirmation if there's an existing button
						if _, err := buttons.GetButton(key); err == nil {
							m.showPasteConfirmation = true
						} else {
							// No existing button, no need to confirm
							_, kv := natsconn.GetNATSConn()
							if err := kv.Delete(key); err == nil {
								kv.Delete(key + ".buffer")
								log.Debug().Str("blanked_button", m.selectedPosition).Msg("Button blanked")
							}
						}
					} else {
						// Normal paste logic for non-blank buttons
						_, kv := natsconn.GetNATSConn()
						if _, err := buttons.GetButton(key); err == nil {
							m.showPasteConfirmation = true
						} else {
							if _, err := kv.Put(key, []byte(m.buttonClipboard)); err == nil {
								log.Debug().Str("pasted_button", m.selectedPosition).Msg("Button pasted")
							}
						}
					}
				}
			case "s": // Start swap mode
				if m.currentDevice != "None" {
					m.swapMode = true
					m.swapSourceButton = m.selectedPosition
					log.Debug().Str("source_button", m.swapSourceButton).Msg("Started button swap")
				}
			case "esc":
				if m.swapMode {
					m.swapMode = false
					m.swapSourceButton = ""
					return m, nil
				}
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

	if m.showDeleteConfirmation {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y", "enter":
				key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
					m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.selectedPosition)
				_, kv := natsconn.GetNATSConn()
				if err := kv.Delete(key); err == nil {
					log.Debug().Str("deleted_button", m.selectedPosition).Msg("Button deleted")
				}
				m.showDeleteConfirmation = false
				return m, nil
			case "n", "esc":
				m.showDeleteConfirmation = false
				return m, nil
			}
			return m, nil // Ignore all other keys when delete dialog is open
		}
	}

	if m.showPasteConfirmation {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y", "enter":
				key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
					m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.selectedPosition)
				_, kv := natsconn.GetNATSConn()

				// If we're pasting a blank button, delete instead of put
				if m.buttonClipboard == "{}" {
					if err := kv.Delete(key); err == nil {
						kv.Delete(key + ".buffer")
						log.Debug().Str("blanked_button", m.selectedPosition).Msg("Button blanked")
					}
				} else {
					if _, err := kv.Put(key, []byte(m.buttonClipboard)); err == nil {
						log.Debug().Str("pasted_button", m.selectedPosition).Msg("Button pasted")
					}
				}
				m.showPasteConfirmation = false
				return m, nil
			case "n", "esc":
				m.showPasteConfirmation = false
				return m, nil
			}
			return m, nil // Ignore all other keys when paste dialog is open
		}
	}

	if m.swapMode {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter":
				if m.swapSourceButton != m.selectedPosition {
					sourceKey := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
						m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.swapSourceButton)
					targetKey := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
						m.currentInstance, m.currentDevice, m.currentProfile, m.currentPage, m.selectedPosition)

					_, kv := natsconn.GetNATSConn()
					sourceButton, err := buttons.GetButton(sourceKey)
					targetButton, err2 := buttons.GetButton(targetKey)

					// Swap the buttons
					if err == nil {
						if data, err := json.Marshal(sourceButton); err == nil {
							kv.Put(targetKey, data)
							// Delete the old buffer to force an update
							kv.Delete(targetKey + ".buffer")
						}
					} else {
						kv.Delete(targetKey)
						kv.Delete(targetKey + ".buffer")
					}

					if err2 == nil {
						if data, err := json.Marshal(targetButton); err == nil {
							kv.Put(sourceKey, data)
							// Delete the old buffer to force an update
							kv.Delete(sourceKey + ".buffer")
						}
					} else {
						kv.Delete(sourceKey)
						kv.Delete(sourceKey + ".buffer")
					}

					log.Debug().
						Str("source", m.swapSourceButton).
						Str("target", m.selectedPosition).
						Msg("Swapped buttons")
				}
				m.swapMode = false
				m.swapSourceButton = ""
				return m, nil
			case "esc":
				m.swapMode = false
				m.swapSourceButton = ""
				return m, nil
			}
		}
	}

	return m, cmd
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

func init() {
	// Register default plugins
	actions.RegisterPlugin(&browser.BrowserPlugin{})
	actions.RegisterPlugin(&command.CommandPlugin{})
	actions.RegisterPlugin(&keyboard.KeyboardPlugin{})
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

	if m.showDeleteConfirmation {
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")). // Red border for delete
			Padding(1, 2).
			Align(lipgloss.Center).
			Width(50).
			MarginLeft((m.width - 54) / 2). // Account for border and padding
			MarginTop((m.height - 6) / 2)   // Roughly center vertically

		return style.Render(
			"Are you sure you want to delete this button?\n\n" +
				"Press 'y' to confirm, 'n' or ESC to cancel",
		)
	}

	if m.showPasteConfirmation {
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")). // Red border for paste
			Padding(1, 2).
			Align(lipgloss.Center).
			Width(50).
			Align(lipgloss.Center).
			MarginLeft((m.width - 54) / 2). // Account for border and padding
			MarginTop((m.height - 6) / 2)   // Roughly center vertically

		return style.Render(
			"Do you want to paste the button here?\n\n" +
				"Press 'y' to confirm, 'n' or ESC to cancel",
		)
	}

	// Create device view
	deviceView := NewDeviceView(m.currentDevice, m.selectedPosition, m.swapMode, m.swapSourceButton)

	// In the View method, before the JoinHorizontal:
	mainContent := fmt.Sprintf(`
Current Instance: %s
Current Device: %s
%s

[i] to change the instance
[d] to change the device
%s
[x] to delete the selected button
[c] to copy button
[v] to paste button
[s] to swap buttons
[q] to quit
%s`,
		m.currentInstance,
		m.currentDevice,
		// Only show profile/page info if device is selected
		func() string {
			if m.currentDevice != "None" {
				return fmt.Sprintf(
					"Current Profile: %s (%s)\nCurrent Page: %s\nCurrent Button: %s",
					m.currentProfileName, m.currentProfile, m.currentPage, m.currentButton,
				)
			}
			return ""
		}(),
		// Only show profile/page controls if device is selected
		func() string {
			if m.currentDevice != "None" {
				return "[p] to change the profile\n[g] to change the page"
			}
			return ""
		}(),
		func() string {
			if m.swapMode {
				return fmt.Sprintf("\n\nSwap Mode: Button %s selected\nUse arrow keys and press ENTER to swap with another button, ESC to cancel",
					m.swapSourceButton)
			}
			return ""
		}(),
	)

	// If we have a device selected, show the device view next to the main content
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainContent,
		"    ", // Add some spacing
		deviceView.View(),
	)
}
