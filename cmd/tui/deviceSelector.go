package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Item represents a single item in the device list
type Item string

// FilterValue returns the value used for filtering
func (i Item) FilterValue() string {
	return string(i) // Return the string value of the Item
}

// Styling for list items and selection
var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

// DeviceSelector represents the state of the device picker overlay
type DeviceSelector struct {
	list list.Model
}

// Devices list (note: using Item type here)
var devices = []list.Item{
	Item("Device 1"),
	Item("Device 2"),
	Item("Device 3"),
}

// NewDeviceSelector creates a new instance of DeviceSelector
func NewDeviceSelector() DeviceSelector {
	deviceList := list.New(devices, itemDelegate{}, 20, 14)
	deviceList.Title = "Select a Device"
	deviceList.SetShowStatusBar(false)
	deviceList.SetFilteringEnabled(false)

	return DeviceSelector{
		list: deviceList,
	}
}

// Init is required by the tea.Model interface, but not used in this case
func (ds DeviceSelector) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates the state of the device selector
func (ds DeviceSelector) Update(msg tea.Msg) (DeviceSelector, tea.Cmd) {
	var cmd tea.Cmd
	ds.list, cmd = ds.list.Update(msg)
	return ds, cmd
}

// View renders the device selector as a string, including the list of devices
func (ds DeviceSelector) View() string {
	return popupStyle.Render(ds.list.View())
}

// itemDelegate handles rendering of each item in the device selection list
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprint(i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// Styling for the overlay (popup)
var popupStyle = lipgloss.NewStyle()
