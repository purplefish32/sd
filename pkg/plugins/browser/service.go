package browser

var PluginNamespace = "sd.plugin.browser"

// BrowserPlugin represents the browser plugin
type BrowserPlugin struct{}

// Name returns the name of the plugin
func (b *BrowserPlugin) Name() string {
	return "browser"
}

// Subscribe sets up the NATS subscription for this plugin
func (c *BrowserPlugin) Init() {
	OpenSubscriber()
}

