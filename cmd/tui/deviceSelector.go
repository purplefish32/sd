package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Item represents a single item in the device list
type Device string

// FilterValue returns the value used for filtering
func (d Device) FilterValue() string {
	return string(d) // Return the string value of the Item
}

// DeviceSelector represents the state of the device selector overlay
type DeviceSelector struct {
	list         list.Model
	width        int
	selectedItem string
}

// Devices list (note: using Item type here)
var devices = []list.Item{
	Device("CL50K2A03427"),
	Device("FL14L1A03452"),
	Device("A00WA32111VSU7"),
}

// NewDeviceSelector creates a new instance of DeviceSelector
func NewDeviceSelector() DeviceSelector {
	deviceList := list.New(devices, deviceDelegate{}, 20, 14)
	deviceList.Title = TitleStyle.Render("Select a Device")
	deviceList.SetShowStatusBar(false)
	deviceList.SetFilteringEnabled(false)
	deviceList.Styles.Title = TitleStyle

	return DeviceSelector{
		list: deviceList,
	}
}

// Init is required by the tea.Model interface, but not used in this case
func (s DeviceSelector) Init() tea.Cmd {
	return nil
}

func (s DeviceSelector) SelectedDevice() string {
	return s.selectedItem
}

// Define a custom message type for when a device is selected
type DeviceSelected string

// Update processes messages and updates the state of the device selector
// Update processes messages and updates the state of the device selector
func (s *DeviceSelector) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return func() tea.Msg {
				return DeviceSelected("")
			}
		case "enter":
			// When a device is selected, store it and return
			selected, ok := s.list.SelectedItem().(Device)
			if ok {
				s.selectedItem = string(selected)
			}
			// Return a command that signals the parent model to handle the selected device
			return func() tea.Msg {
				return DeviceSelected(s.selectedItem) // Custom message to signal parent model
			}
		}
	}
	s.list, cmd = s.list.Update(msg)
	return cmd
}

// View renders the device selector as a string, including the list of devices
func (s DeviceSelector) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(s.width - 4)

	// Update list styles
	s.list.Styles.Title = TitleStyle
	s.list.SetSize(s.width-8, 14) // Account for borders and padding

	return style.Render("Select Device:\n\n" + s.list.View())
}

// deviceDelegate handles rendering of each item in the device selection list
type deviceDelegate struct{}

func (d deviceDelegate) Height() int                             { return 1 }
func (d deviceDelegate) Spacing() int                            { return 0 }
func (d deviceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d deviceDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Device)
	if !ok {
		return
	}

	str := fmt.Sprint(i)
	if index == m.Index() {
		str = "> " + str // Add a ">" prefix for the selected item
	}

	fmt.Fprint(w, str)
}
