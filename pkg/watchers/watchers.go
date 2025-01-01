package watchers

import (
	"sd/pkg/streamdeck"
	"time"

	"github.com/karalabe/hid"
	"github.com/rs/zerolog/log"
)

var knownDevices = make(map[string]bool)

func WatchForStreamDecks(instanceID string) {
	log.Info().Interface("knownDevices", knownDevices).Msg("knownDevices")
	for {
		devices := hid.Enumerate(streamdeck.ElgatoVendorID, 0)

		currentDevices := make(map[string]bool)

		// Check for new devices.
		for _, device := range devices {
			deviceKey := device.Serial

			currentDevices[deviceKey] = true

			if !knownDevices[deviceKey] {
				log.Info().Interface("device", device).Msg("Stream Deck connected")

				openDevice, err := device.Open()
				if err != nil {
					log.Error().Err(err).Str("device", deviceKey).Msg("Failed to open device")
					continue
				}

				sd := streamdeck.New(instanceID, openDevice)
				go sd.Init()

				knownDevices[deviceKey] = true
			}
		}

		// Check for removed devices.
		for deviceKey := range knownDevices {
			if !currentDevices[deviceKey] {
				log.Info().Str("device", deviceKey).Msg("Stream Deck disconnected")
				delete(knownDevices, deviceKey)

				// Perform cleanup for removed devices.
				streamdeck.RemoveDevice(deviceKey)
			}
		}

		// Sleep for a short interval before checking again.
		time.Sleep(time.Second)
	}
}
