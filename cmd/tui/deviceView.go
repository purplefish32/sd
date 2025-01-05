package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type DeviceView struct {
	deviceID string
}

func NewDeviceView(deviceID string) DeviceView {
	return DeviceView{
		deviceID: deviceID,
	}
}

func (d DeviceView) View() string {
	if d.deviceID == "None" {
		return ""
	}

	// Create styles for the buttons
	buttonStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Align(lipgloss.Center).
		Width(6).
		Height(3)

	var grid []string

	// Check device type based on serial number pattern
	if strings.HasPrefix(d.deviceID, "CL") { // Stream Deck Plus
		const (
			numRows = 2
			cols    = 4
		)
		for r := 0; r < numRows; r++ {
			var row []string
			for c := 0; c < cols; c++ {
				button := buttonStyle.Render(fmt.Sprintf("%d", r*cols+c+1))
				row = append(row, button)
			}
			grid = append(grid, lipgloss.JoinHorizontal(lipgloss.Top, row...))
		}
	} else if strings.HasPrefix(d.deviceID, "FL") { // Stream Deck Pedal
		const numButtons = 3
		var row []string
		for i := 0; i < numButtons; i++ {
			button := buttonStyle.Render(fmt.Sprintf("%d", i+1))
			row = append(row, button)
		}
		grid = append(grid, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	} else { // Stream Deck XL
		const (
			numRows = 4
			cols    = 8
		)
		for r := 0; r < numRows; r++ {
			var row []string
			for c := 0; c < cols; c++ {
				button := buttonStyle.Render(fmt.Sprintf("%d", r*cols+c+1))
				row = append(row, button)
			}
			grid = append(grid, lipgloss.JoinHorizontal(lipgloss.Top, row...))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, grid...)
}
