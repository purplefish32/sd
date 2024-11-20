package core

import (
	"sync"

	"github.com/nats-io/nats.go"
)

// Plugin interface that all plugins must implement
type Plugin interface {
	Name() string         // Returns the plugin's name
	Subscribe(nc *nats.Conn) error // Sets up the plugin's NATS subscriptions
}

// PluginRegistry manages the list of registered plugins
type PluginRegistry struct {
	mu      sync.Mutex
	plugins map[string]Plugin
}

// NewPluginRegistry creates and initializes a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry
func (r *PluginRegistry) Register(plugin Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[plugin.Name()] = plugin
}

// Get returns a plugin by name
func (r *PluginRegistry) Get(name string) (Plugin, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	plugin, exists := r.plugins[name]
	return plugin, exists
}

// All returns all registered plugins
func (r *PluginRegistry) All() []Plugin {
	r.mu.Lock()
	defer r.mu.Unlock()
	allPlugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		allPlugins = append(allPlugins, plugin)
	}
	return allPlugins
}