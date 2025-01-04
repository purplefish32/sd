package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Item represents a single item in the instance list
type Instance string

// FilterValue returns the value used for filtering
func (i Instance) FilterValue() string {
	return string(i) // Return the string value of the Item
}

// InstanceSelector represents the state of the instance picker overlay
type InstanceSelector struct {
	list         list.Model
	selectedItem string
}

// Instances list (note: using Item type here)
var instances = []list.Item{
	Instance("Instance 1"),
	Instance("Instance 2"),
	Instance("Instance 3"),
}

// NewInstanceSelector creates a new instance of InstanceSelector
func NewInstanceSelector() InstanceSelector {
	instanceList := list.New(instances, instanceDelegate{}, 20, 14)
	instanceList.Title = "Select an Instance"
	instanceList.SetShowStatusBar(false)
	instanceList.SetFilteringEnabled(false)

	return InstanceSelector{
		list: instanceList,
	}
}

// Init is required by the tea.Model interface, but not used in this case
func (s InstanceSelector) Init() tea.Cmd {
	return nil
}

func (s InstanceSelector) SelectedInstance() string {
	return s.selectedItem
}

// Define a custom message type for when a device is selected
type InstanceSelected string

// Update processes messages and updates the state of the instance selector
func (s *InstanceSelector) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// When an instance is selected, store it
			selected, ok := s.list.SelectedItem().(Instance)
			if ok {
				s.selectedItem = string(selected)
			}
			// Return a command that signals the parent model to handle the selected device
			return func() tea.Msg {
				return InstanceSelected(s.selectedItem) // Custom message to signal parent model
			}
		}
	}
	s.list, cmd = s.list.Update(msg)
	return cmd
}

// View renders the instance selector as a string, including the list of instances
func (s InstanceSelector) View() string {
	return s.list.View()
}

// instanceDelegate handles rendering of each item in the instance selection list
type instanceDelegate struct{}

func (d instanceDelegate) Height() int                             { return 1 }
func (d instanceDelegate) Spacing() int                            { return 0 }
func (d instanceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d instanceDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Instance)
	if !ok {
		return
	}

	str := fmt.Sprint(i)
	if index == m.Index() {
		str = "> " + str // Add a ">" prefix for the selected item
	}

	fmt.Fprint(w, str)
}
