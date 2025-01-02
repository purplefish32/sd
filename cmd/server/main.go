package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"sd/pkg/core"
	"sd/pkg/env"
	"sd/pkg/instance"
	"sd/pkg/plugins/browser"
	"sd/pkg/plugins/command"
	"sd/pkg/plugins/keyboard"
	"sd/pkg/watchers"
)

func main() {
	// Set global time format for logger.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Configure the global logger.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	log.Info().Msg("Starting application")

	// Retrieve or create the instance UUID.
	instanceID := instance.GetOrCreateInstanceUUID()

	// Load environment variables
	env.LoadEnv()

	// Register plugins.
	registry := core.NewPluginRegistry()
	registry.Register(&browser.BrowserPlugin{})
	registry.Register(&command.CommandPlugin{})
	registry.Register(&keyboard.KeyboardPlugin{})

	// Initialize plugins.
	for _, plugin := range registry.All() {
		log.Info().Str("plugin", plugin.Name()).Msg("Registering plugin")
		plugin.Init()
		log.Info().Str("plugin", plugin.Name()).Msg("Plugin subscribed successfully")
	}

	// Start watching Stream Deck devices.
	go watchers.WatchStreamDecks(instanceID)

	// Keep the main program running.
	select {}
}
