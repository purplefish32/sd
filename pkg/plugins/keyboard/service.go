package keyboard

import (
	"encoding/json"
	"sd/pkg/actions"

	"github.com/go-vgo/robotgo"
)

var pluginNamespace = "sd.plugin.keyboard"

// KeyboardPlugin represents the keyboard plugin.
type KeyboardPlugin struct{}

type TypeConfig struct {
	Text string `json:"text"`
}

// Name returns the name of the plugin.
func (k *KeyboardPlugin) Name() string {
	return "keyboard"
}

// Subscribe sets up the NATS subscription for this plugin.
func (k *KeyboardPlugin) Init() {
	SubscribeActionType(pluginNamespace)
}

func (k *KeyboardPlugin) GetActionTypes() []actions.ActionType {
	return []actions.ActionType{
		"type",
	}
}

func (k *KeyboardPlugin) ValidateConfig(actionType actions.ActionType, config json.RawMessage) error {
	var cfg TypeConfig
	return json.Unmarshal(config, &cfg)
}

func (k *KeyboardPlugin) ExecuteAction(actionType actions.ActionType, config json.RawMessage) error {
	var cfg TypeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return err
	}
	robotgo.TypeStr(cfg.Text)
	return nil
}
