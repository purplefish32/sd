package main

import (
	"fmt"
	"io"
	"strings"

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
	list list.Model
}

// Instances list (note: using Item type here)
var instances = []list.Item{
	Item("Instance 1"),
	Item("Instance 2"),
	Item("Instance 3"),
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
func (ds InstanceSelector) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates the state of the instance selector
func (ds InstanceSelector) Update(msg tea.Msg) (InstanceSelector, tea.Cmd) {
	var cmd tea.Cmd
	ds.list, cmd = ds.list.Update(msg)
	return ds, cmd
}

// View renders the instance selector as a string, including the list of instances
func (ds InstanceSelector) View() string {
	return popupStyle.Render(ds.list.View())
}

// instanceDelegate handles rendering of each item in the instance selection list
type instanceDelegate struct{}

func (d instanceDelegate) Height() int                             { return 1 }
func (d instanceDelegate) Spacing() int                            { return 0 }
func (d instanceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d instanceDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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
