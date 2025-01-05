package actions

import (
	"encoding/json"
	"fmt"
	"sd/pkg/buttons"
)

type ActionType string

// Plugin interface that all plugins must implement
type Plugin interface {
	Name() string
	Init()
	GetActionTypes() []ActionType
	ValidateConfig(actionType ActionType, config json.RawMessage) error
	ExecuteAction(actionType ActionType, config json.RawMessage) error
}

type Action struct {
	PluginName string          `json:"plugin_name"`
	Type       ActionType      `json:"type"`
	Config     json.RawMessage `json:"config"`
}

// Registry to keep track of available plugins
type Registry struct {
	plugins map[string]Plugin
}

var globalRegistry = &Registry{
	plugins: make(map[string]Plugin),
}

func RegisterPlugin(plugin Plugin) error {
	name := plugin.Name()
	if _, exists := globalRegistry.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}
	globalRegistry.plugins[name] = plugin
	plugin.Init()
	return nil
}

func GetPlugin(name string) (Plugin, error) {
	plugin, exists := globalRegistry.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return plugin, nil
}

func ExecuteAction(action *Action) error {
	plugin, err := GetPlugin(action.PluginName)
	if err != nil {
		return err
	}
	return plugin.ExecuteAction(action.Type, action.Config)
}

type State struct {
	Id        string `json:"id"`
	ImagePath string `json:"imagePath"`
}

type ActionInstance struct {
	UUID     string          `json:"uuid"`
	Settings any             `json:"settings"`
	State    string          `json:"state"`
	States   []buttons.State `json:"states"`
	Title    string          `json:"title"`
}

func GetRegisteredPlugins() map[string]Plugin {
	return globalRegistry.plugins
}

// ExecuteButtonAction executes the action configured for a button
func ExecuteButtonAction(button buttons.Button) error {
	switch button.UUID {
	case "sd.plugin.browser.open":
		plugin, _ := GetPlugin("browser")
		config, _ := json.Marshal(button.Settings)
		return plugin.ExecuteAction("open_url", config)

	case "sd.plugin.keyboard.type":
		plugin, _ := GetPlugin("keyboard")
		config, _ := json.Marshal(button.Settings)
		return plugin.ExecuteAction("type", config)

	case "sd.plugin.command.exec":
		plugin, _ := GetPlugin("command")
		config, _ := json.Marshal(button.Settings)
		return plugin.ExecuteAction("exec", config)
	}
	return fmt.Errorf("unknown action type: %s", button.UUID)
}
