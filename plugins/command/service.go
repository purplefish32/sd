package command

var pluginNamespace = "sd.plugin.command"

// CommandPlugin represents the command plugin.
type CommandPlugin struct{}

// Name returns the name of the plugin.
func (c *CommandPlugin) Name() string {
	return "command"
}

// Subscribe sets up the NATS subscription for this plugin.
func (c *CommandPlugin) Init() {
	SubscribeActionExec(pluginNamespace)
}