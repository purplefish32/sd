package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"sd/pkg/actions"
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

const (
	stateEditing = iota
	stateSelectingPlugin
	stateSelectingAction
	stateConfiguringAction
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
	textarea            textarea.Model
	buttonNum           string
	showEditor          bool
	instanceID          string
	deviceID            string
	profileID           string
	pageID              string
	jsonValid           bool
	jsonError           string
	width               int
	showActionSelector  bool
	selectedPlugin      string
	selectedAction      string
	actionConfig        map[string]interface{}
	currentState        int
	availablePlugins    []string
	availableActions    []actions.ActionType
	selectedPluginIndex int
	selectedActionIndex int
	configInput         string
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

	switch e.currentState {
	case stateEditing:
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
			case "a":
				e.currentState = stateSelectingPlugin
				e.availablePlugins = make([]string, 0, len(actions.GetRegisteredPlugins()))
				for name := range actions.GetRegisteredPlugins() {
					e.availablePlugins = append(e.availablePlugins, name)
				}
				return e, nil
			}
		}
	case stateSelectingPlugin:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if e.selectedPluginIndex > 0 {
					e.selectedPluginIndex--
				}
			case "down", "j":
				if e.selectedPluginIndex < len(e.availablePlugins)-1 {
					e.selectedPluginIndex++
				}
			case "esc":
				e.currentState = stateEditing
			case "enter":
				if len(e.availablePlugins) > 0 {
					e.selectedPlugin = e.availablePlugins[e.selectedPluginIndex]
					plugin, _ := actions.GetPlugin(e.selectedPlugin)
					e.availableActions = plugin.GetActionTypes()
					e.selectedActionIndex = 0
					e.currentState = stateSelectingAction
				}
			}
		}
	case stateSelectingAction:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if e.selectedActionIndex > 0 {
					e.selectedActionIndex--
				}
			case "down", "j":
				if e.selectedActionIndex < len(e.availableActions)-1 {
					e.selectedActionIndex++
				}
			case "esc":
				e.currentState = stateSelectingPlugin
			case "enter":
				if len(e.availableActions) > 0 {
					e.selectedAction = string(e.availableActions[e.selectedActionIndex])
					e.actionConfig = make(map[string]interface{})
					e.currentState = stateConfiguringAction
				}
			}
		}
	case stateConfiguringAction:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				e.currentState = stateSelectingAction
			case "enter":
				var button buttons.Button
				switch e.selectedPlugin {
				case "browser":
					button = buttons.Button{
						UUID:     "sd.plugin.browser.open_url",
						Settings: buttons.Settings{URL: e.configInput},
						States: []buttons.State{
							{Id: "0", ImagePath: "/home/donovan/.config/sd/buttons/google.png"},
						},
						State: "0",
						Title: "",
					}
				case "keyboard":
					button = buttons.Button{
						UUID:     "sd.plugin.keyboard.type",
						Settings: buttons.Settings{Text: e.configInput},
						States: []buttons.State{
							{Id: "0", ImagePath: "/home/donovan/.config/sd/buttons/keyboard.png"},
						},
						State: "0",
						Title: "",
					}
				case "command":
					button = buttons.Button{
						UUID:     "sd.plugin.command.exec",
						Settings: buttons.Settings{Command: e.configInput},
						States: []buttons.State{
							{Id: "0", ImagePath: "/home/donovan/.config/sd/buttons/terminal.png"},
						},
						State: "0",
						Title: "",
					}
				}

				if jsonData, err := json.MarshalIndent(button, "", "  "); err == nil {
					e.textarea.SetValue(string(jsonData))
				}
				e.currentState = stateEditing
			default:
				if msg.Type == tea.KeyRunes {
					e.configInput += msg.String()
				} else if msg.Type == tea.KeyBackspace && len(e.configInput) > 0 {
					e.configInput = e.configInput[:len(e.configInput)-1]
				}
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

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(e.width - 4).
		Align(lipgloss.Center)

	var content string
	switch e.currentState {
	case stateSelectingPlugin:
		var pluginList strings.Builder
		pluginList.WriteString("Select a Plugin:\n\n")
		for i, plugin := range e.availablePlugins {
			prefix := "  "
			if i == e.selectedPluginIndex {
				prefix = "> "
			}
			pluginList.WriteString(fmt.Sprintf("%s%s\n", prefix, plugin))
		}
		content = pluginList.String()
		return style.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			"Select Plugin",
			"",
			content,
			"",
			"Press ENTER to select, ESC to cancel",
		))

	case stateSelectingAction:
		var actionList strings.Builder
		actionList.WriteString(fmt.Sprintf("Plugin: %s\nSelect an Action:\n\n", e.selectedPlugin))
		for i, action := range e.availableActions {
			prefix := "  "
			if i == e.selectedActionIndex {
				prefix = "> "
			}
			actionList.WriteString(fmt.Sprintf("%s%s\n", prefix, action))
		}
		content = actionList.String()
		return style.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			"Select Action",
			"",
			content,
			"",
			"Press ENTER to select, ESC to cancel",
		))

	case stateConfiguringAction:
		var configForm strings.Builder
		configForm.WriteString(fmt.Sprintf("Configure %s - %s\n\n", e.selectedPlugin, e.selectedAction))

		switch e.selectedPlugin {
		case "browser":
			configForm.WriteString("Enter URL: ")
			configForm.WriteString(e.configInput)
		case "keyboard":
			configForm.WriteString("Enter text to type: ")
			configForm.WriteString(e.configInput)
		case "command":
			configForm.WriteString("Enter command: ")
			configForm.WriteString(e.configInput)
		}

		configForm.WriteString("\n\nPress ENTER to save, ESC to cancel")

		return style.Render(configForm.String())

	case stateEditing:
		leftColumn := lipgloss.JoinVertical(
			lipgloss.Left,
			"Editor (press 'a' for actions)",
			"",
			e.textarea.View(),
		)
		rightColumn := lipgloss.JoinVertical(
			lipgloss.Left,
			"Preview",
			"",
			highlightJSON(e.textarea.Value()),
		)

		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftColumn,
			"  │  ", // Separator
			rightColumn,
		)
	}

	return style.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("Button %s Configuration", e.buttonNum),
		"",
		content,
		"",
		"Press CTRL+S to save, ESC to close",
	))
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
		e.textarea.SetValue("{}")
		return
	}

	jsonData, err := json.MarshalIndent(button, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal button data")
		return
	}

	e.textarea.SetValue(string(jsonData))
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
		return fmt.Errorf("failed to unmarshal button: %w", err)
	}

	_, kv := natsconn.GetNATSConn()
	data, err := json.Marshal(button)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal button data")
		return fmt.Errorf("failed to marshal button: %w", err)
	}

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
