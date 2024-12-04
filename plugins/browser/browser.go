package browser

import (
	"sd/plugins/browser/actions"
)

var PluginNamespace = "sd.plugin.core.browser"

// BrowserPlugin represents the browser plugin
type BrowserPlugin struct{}

// Name returns the name of the plugin
func (b *BrowserPlugin) Name() string {
	return "command"
}

// Subscribe sets up the NATS subscription for this plugin
func (c *BrowserPlugin) Init() {
	actions.SubscribeActionOpen(PluginNamespace)
}

