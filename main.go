package main

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"sd/core"
	"sd/plugins/browser"
	"sd/plugins/command"
	"sd/plugins/keyboard"
	"sd/streamdeck"
	streamdeckPedal "sd/streamdeck/pedal"
	streamdeckXl "sd/streamdeck/xl"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/karalabe/hid"
)

func main() {
	// Set global time format for logger.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Configure the global logger.
    log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	log.Info().Msg("Starting application")

	// Retrieve or create the instance UUID.
	instanceID := getOrCreateUUID()

	// Load the .env file.
	err := godotenv.Load()

	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
		os.Exit(1) // Explicitly terminate the program.
	}

	// Register plugins.
	registry := core.NewPluginRegistry()
	registry.Register(&browser.BrowserPlugin{})
	registry.Register(&command.CommandPlugin{})
	registry.Register(&keyboard.KeyboardPlugin{})

	// Initialize plugins.
	for _, plugin := range registry.All() {
		log.Info().Str("plugin", plugin.Name()).Msg("Registering plugin")
		plugin.Init();
        log.Info().Str("plugin", plugin.Name()).Msg("Plugin subscribed successfully")
	}

	// Define devices.
	deviceTypes := []struct {
		Name      string
		VendorID  uint16
		ProductID uint16
		Init      func(instanceID string, device *hid.Device)

	}{
		{"Stream Deck XL", streamdeck.VendorID, streamdeckXl.ProductID, streamdeckXl.Init},
		{"Stream Deck Pedal", streamdeck.VendorID, streamdeckPedal.ProductID, streamdeckPedal.Init},
		//{"Stream Deck +", streamdeck.VendorID, streamdeckPlus.ProductID, streamdeckPlus.Initialize},
	}

	// Process each device type.
	for _, deviceType := range deviceTypes {
		go func(dt struct {
			Name       string
			VendorID   uint16
			ProductID  uint16
			Init func(instanceID string, device *hid.Device)
		}) {
			// Find the devices.
			hidDevices := hid.Enumerate(dt.VendorID, dt.ProductID)

			if len(hidDevices) == 0 {
				log.Warn().Str("device_name", dt.Name).Msg("Device not found")
				return
			}

			// Process each device of this type.
			for i, hidDeviceInfo := range hidDevices {
				go func(deviceIndex int, deviceInfo hid.DeviceInfo) {

					// Open the device.
					hidDevice, err := deviceInfo.Open()

					if err != nil {
						log.Error().Err(err).Str("device_name", dt.Name).Msg("Failed to open Device")
						return
					}

					// Initialize the device.
					dt.Init(instanceID, hidDevice)
				}(i, hidDeviceInfo)
			}
		}(deviceType)
	}

	// Keep the main program running.
	select {}
}

func getOrCreateUUID() string {

	// Use a directory in the user's home folder
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving user home directory")
	}

	uuidDir := filepath.Join(homeDir, ".config/sd")
	uuidFilePath := filepath.Join(uuidDir, "instance-id")

	// Ensure the directory exists
	if _, err := os.Stat(uuidDir); os.IsNotExist(err) {
		err := os.MkdirAll(uuidDir, 0755) // Create the directory

		if err != nil {
			log.Fatal().Err(err).Str("uuidDir", uuidDir).Msg("Error creating directory")
			os.Exit(1)
		}
	}

	// Check if the UUID file exists
	if _, err := os.Stat(uuidFilePath); err == nil {

		// Read the existing UUID
		data, err := os.ReadFile(uuidFilePath)

		uuid := string(data)

		if err != nil {
			log.Fatal().Err(err).Msg("Error reading UUID file")
			os.Exit(1)
		}

		log.Info().Str("uuid", uuid).Msg("UUID file exists")

		return uuid
	}

	// Generate a new UUID
	id := uuid.New()
	uuid := id.String()

	// Save the UUID to the file
	err = os.WriteFile(uuidFilePath, []byte(uuid), 0600)

	if err != nil {
		log.Fatal().Err(err).Msg("Error saving UUID to file")
	}

	log.Info().Str("uuid", uuid).Msg("UUID file created")

	return uuid
}