package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ButtonEditor struct {
	textarea   textarea.Model
	buttonNum  string
	showEditor bool
	width      int
	height     int
}

type EditorClosing struct{}

func NewButtonEditor() ButtonEditor {
	ta := textarea.New()
	ta.Placeholder = "Enter button configuration JSON..."
	ta.Focus()

	return ButtonEditor{
		textarea:   ta,
		showEditor: false,
	}
}

func (e ButtonEditor) Init() tea.Cmd {
	return textarea.Blink
}

func (e *ButtonEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			e.showEditor = false
			return e, func() tea.Msg {
				return EditorClosing{}
			}
		}
	}

	var cmd tea.Cmd
	e.textarea, cmd = e.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return e, tea.Batch(cmds...)
}

func (e ButtonEditor) View() string {
	if !e.showEditor {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center)

	return style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			fmt.Sprintf("Button %s Configuration", e.buttonNum),
			"",
			e.textarea.View(),
			"",
			"Press ESC to close",
		),
	)
}
