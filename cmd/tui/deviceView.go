package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type DeviceView struct {
	deviceID       string
	selectedButton string
}

func NewDeviceView(deviceID string, selectedButton string) DeviceView {
	return DeviceView{
		deviceID:       deviceID,
		selectedButton: selectedButton,
	}
}

func (d DeviceView) View() string {
	if d.deviceID == "None" {
		return ""
	}

	// Create styles for the buttons
	buttonStyle := ListStyle.Copy().
		Width(6).
		Height(3).
		Align(lipgloss.Center)

	selectedButtonStyle := buttonStyle.Copy().
		BorderForeground(lipgloss.Color("205"))

	var grid []string

	// Check device type based on serial number pattern
	if strings.HasPrefix(d.deviceID, "A0") { // Stream Deck Plus
		const (
			numRows = 2
			cols    = 4
		)
		for r := 0; r < numRows; r++ {
			var row []string
			for c := 0; c < cols; c++ {
				buttonNum := r*cols + c + 1
				style := buttonStyle
				if fmt.Sprintf("%d", buttonNum) == d.selectedButton {
					style = selectedButtonStyle
				}
				button := style.Render(fmt.Sprintf("%d", buttonNum))
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
	} else if strings.HasPrefix(d.deviceID, "CL") { // Stream Deck XL
		const (
			numRows = 4
			cols    = 8
		)
		for r := 0; r < numRows; r++ {
			var row []string
			for c := 0; c < cols; c++ {
				buttonNum := r*cols + c + 1
				style := buttonStyle
				if fmt.Sprintf("%d", buttonNum) == d.selectedButton {
					style = selectedButtonStyle
				}
				button := style.Render(fmt.Sprintf("%d", buttonNum))
				row = append(row, button)
			}
			grid = append(grid, lipgloss.JoinHorizontal(lipgloss.Top, row...))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, grid...)
}
