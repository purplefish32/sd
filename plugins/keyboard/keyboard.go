package keyboard

import "sd/plugins/keyboard/actions"

var pluginNamespace = "sd.plugin.core.keyboard"

// KeyboardPlugin represents the keyboard plugin.
type KeyboardPlugin struct{}

// Name returns the name of the plugin.
func (k *KeyboardPlugin) Name() string {
	return "keyboard"
}

// Subscribe sets up the NATS subscription for this plugin.
func (k *KeyboardPlugin) Init() {
	actions.SubscribeActionType(pluginNamespace)
}