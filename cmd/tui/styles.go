package main

import "github.com/charmbracelet/lipgloss"

var (
	// Common styles used across components
	ListStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true).
			Padding(0, 1)

	SelectedBorderStyle = lipgloss.NewStyle().
				BorderForeground(lipgloss.Color("205")) // Bright magenta for selected items
)
