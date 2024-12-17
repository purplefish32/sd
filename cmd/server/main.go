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
	"sd/pkg/streamdeck"
	"sd/pkg/watchers"

	"github.com/karalabe/hid"
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

	go watchers.WatchForButtonChanges()

	// Process each device type.
	for _, streamDeckType := range streamdeck.StreamDeckTypes {
		// Find all locally connected Stream Decks of the given type.
		devices := hid.Enumerate(streamdeck.ElgatoVendorID, streamDeckType.ProductID)

		if len(devices) == 0 {
			log.Warn().Msg("No Stream Decks connected")
			return
		}

		// Process each Stream Deck type.
		for _, device := range devices {
			log.Info().Interface("device", device).Msg("Stream Deck found")

			// Open the device.
			openDevice, err := hid.DeviceInfo.Open(device)

			if err != nil {
				log.Error().Err(err).Str("device_name", streamDeckType.Name).Msg("Failed to open Device")
				return
			}

			log.Info().Interface("stream_deck", openDevice).Msg("Stream Deck opened")

			// Initialize the device.
			sd := streamdeck.New(instanceID, openDevice)
			go sd.Init()
		}
	}

	// Keep the main program running.
	select {}
}
