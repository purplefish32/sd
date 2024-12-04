package command

import "sd/plugins/command/actions"

var pluginNamespace = "sd.plugin.core.command"

// CommandPlugin represents the command plugin.
type CommandPlugin struct{}

// Name returns the name of the plugin.
func (c *CommandPlugin) Name() string {
	return "command"
}

// Subscribe sets up the NATS subscription for this plugin.
func (c *CommandPlugin) Init() {
	actions.SubscribeActionExec(pluginNamespace)
}