package main

import (
	"encoding/json"
	"fmt"

	"sd/pkg/buttons"
	"sd/pkg/natsconn"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type ButtonEditor struct {
	textarea   textarea.Model
	buttonNum  string
	showEditor bool
	width      int
	height     int
	instanceID string
	deviceID   string
	profileID  string
	pageID     string
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
		case "ctrl+s":
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
		// If button doesn't exist, create a new one with defaults
		button = buttons.Button{
			UUID: "sd.plugin.browser.open",
			States: []buttons.State{
				{
					Id:        "0",
					ImagePath: "",
				},
			},
			State: "0",
			Title: "",
		}
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

	e.textarea.SetValue(string(jsonData))
}

func (e *ButtonEditor) SaveButton() error {
	key := fmt.Sprintf("sd/instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
		e.instanceID, e.deviceID, e.profileID, e.pageID, e.buttonNum)

	var button buttons.Button
	if err := json.Unmarshal([]byte(e.textarea.Value()), &button); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	_, kv := natsconn.GetNATSConn()
	data, err := json.Marshal(button)
	if err != nil {
		return fmt.Errorf("failed to marshal button: %w", err)
	}

	// Try to create first, if it fails with KeyExists, then update
	_, err = kv.Create(key, data)
	if err == nats.ErrKeyExists {
		_, err = kv.Put(key, data)
	}

	return err
}
