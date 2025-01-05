package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"sd/pkg/buttons"
	"sd/pkg/natsconn"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

func highlightJSON(input string) string {
	lexer := lexers.Get("json")
	style := styles.Get("monokai")
	formatter := formatters.Get("terminal256")

	iterator, err := lexer.Tokenise(nil, input)
	if err != nil {
		return input
	}

	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return input
	}

	return buf.String()
}

type ButtonEditor struct {
	textarea   textarea.Model
	buttonNum  string
	showEditor bool
	instanceID string
	deviceID   string
	profileID  string
	pageID     string
	jsonValid  bool
	jsonError  string
	width      int
}

type EditorClosing struct{}

func NewButtonEditor(instanceID, deviceID, profileID, pageID string) ButtonEditor {
	ta := textarea.New()
	ta.Placeholder = "Enter button configuration JSON..."
	ta.Focus()

	return ButtonEditor{
		textarea:   ta,
		showEditor: false,
		instanceID: instanceID,
		deviceID:   deviceID,
		profileID:  profileID,
		pageID:     pageID,
		jsonValid:  true,
		width:      120,
	}
}

func (e ButtonEditor) Init() tea.Cmd {
	return textarea.Blink
}

func (e *ButtonEditor) validateJSON() {
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(e.textarea.Value()), &js); err != nil {
		e.jsonValid = false
		e.jsonError = err.Error()
	} else {
		e.jsonValid = true
		e.jsonError = ""
	}
}

func (e *ButtonEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.width = msg.Width
		e.textarea.SetWidth((e.width - 10) / 2)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			e.showEditor = false
			return e, func() tea.Msg {
				return EditorClosing{}
			}
		case "ctrl+s":
			if !e.jsonValid {
				return e, nil
			}
			if err := e.SaveButton(); err != nil {
				log.Error().Err(err).Msg("Failed to save button")
			}
			e.showEditor = false
			return e, func() tea.Msg {
				return EditorClosing{}
			}
		}
	}

	var cmd tea.Cmd
	e.textarea, cmd = e.textarea.Update(msg)
	e.validateJSON()
	cmds = append(cmds, cmd)

	return e, tea.Batch(cmds...)
}

func (e ButtonEditor) View() string {
	if !e.showEditor {
		return ""
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46"))

	if !e.jsonValid {
		statusStyle = statusStyle.
			Foreground(lipgloss.Color("196"))
	}

	status := "JSON: valid"
	if !e.jsonValid {
		status = fmt.Sprintf("JSON Error: %s", e.jsonError)
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(e.width - 4).
		Align(lipgloss.Center)

	// Create the two columns
	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		"Editor",
		"",
		e.textarea.View(),
	)

	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		"Preview",
		"",
		highlightJSON(e.textarea.Value()),
	)

	// Join the columns side by side
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		"  │  ", // Separator
		rightColumn,
	)

	return style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			fmt.Sprintf("Button %s Configuration", e.buttonNum),
			"",
			content,
			"",
			statusStyle.Render(status),
			"Press CTRL+S to save, ESC to close",
		),
	)
}

func (e *ButtonEditor) LoadButton() {
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
		e.instanceID, e.deviceID, e.profileID, e.pageID, e.buttonNum)

	log.Debug().
		Str("key", key).
		Str("instanceID", e.instanceID).
		Str("deviceID", e.deviceID).
		Str("profileID", e.profileID).
		Str("pageID", e.pageID).
		Str("buttonNum", e.buttonNum).
		Msg("Loading button XXX")

	button, err := buttons.GetButton(key)
	if err != nil {
		log.Debug().Err(err).Str("key", key).Msg("Button not found, creating default")
		// If button doesn't exist, show empty JSON
		e.textarea.SetValue("{}")
		return
	}

	// Convert button to JSON
	jsonData, err := json.MarshalIndent(button, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal button data")
		return
	}

	log.Debug().
		Str("jsonData", string(jsonData)).
		Msg("Setting textarea value")

	formattedJSON := string(jsonData)
	e.textarea.SetValue(formattedJSON)
}

func (e *ButtonEditor) SaveButton() error {
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
		e.instanceID, e.deviceID, e.profileID, e.pageID, e.buttonNum)

	log.Debug().
		Str("key", key).
		Str("value", e.textarea.Value()).
		Msg("Attempting to save button")

	var button buttons.Button
	if err := json.Unmarshal([]byte(e.textarea.Value()), &button); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal button JSON")
		return fmt.Errorf("invalid JSON: %w", err)
	}

	_, kv := natsconn.GetNATSConn()
	data, err := json.Marshal(button)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal button data")
		return fmt.Errorf("failed to marshal button: %w", err)
	}

	// Just use Put instead of trying Create first
	revision, err := kv.Put(key, data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save button")
		return err
	}

	log.Info().
		Str("key", key).
		Uint64("revision", revision).
		Msg("Successfully saved button")
	return nil
}
