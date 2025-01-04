package watchers

import (
	"encoding/json"
	"sd/pkg/natsconn"
	"sd/pkg/streamdeck"
	"time"

	"github.com/karalabe/hid"
	"github.com/rs/zerolog/log"
)

// knownDevices should store hid.DeviceInfo instead of bool
var knownDevices = make(map[string]hid.DeviceInfo)

func WatchStreamDecks(instanceID string) {
	// Connect to NATS.
	nc, _ := natsconn.GetNATSConn()

	for {
		devices := hid.Enumerate(streamdeck.ElgatoVendorID, 0)

		currentDevices := make(map[string]hid.DeviceInfo) // Store the entire device object

		// Check for new devices.
		for _, device := range devices {
			deviceKey := device.Serial

			// Store the device in currentDevices map
			currentDevices[deviceKey] = device

			if _, exists := knownDevices[deviceKey]; !exists {
				log.Info().Interface("device", device).Msg("Stream Deck connected")

				// Serialize the entire device object to JSON for connection event
				payload, err := json.Marshal(device)
				if err != nil {
					log.Error().Err(err).Interface("device", device).Msg("Error serializing device")
					continue // Skip this device if serialization fails
				}

				// Publish the device connection message
				nc.Publish("sd.device.connected", payload)

				// Open the device
				openDevice, err := device.Open()

				if err != nil {
					log.Error().Err(err).Str("device", deviceKey).Msg("Failed to open device")
					continue
				}

				sd := streamdeck.New(instanceID, openDevice)
				go sd.Init()

				// Mark device as known by storing full device info
				knownDevices[deviceKey] = device
			}
		}

		// Check for removed devices.
		for deviceKey, device := range knownDevices {
			if _, exists := currentDevices[deviceKey]; !exists {
				log.Info().Str("device", deviceKey).Msg("Stream Deck disconnected")
				delete(knownDevices, deviceKey)

				// Serialize the full device object to JSON for disconnection event
				payload, err := json.Marshal(device)
				if err != nil {
					log.Error().Err(err).Interface("device", device).Msg("Error serializing device")
					continue // Skip this device if serialization fails
				}

				// Publish the device disconnection message
				nc.Publish("sd.device.disconnected", payload)

				// Perform cleanup for removed devices.
				streamdeck.RemoveDevice(deviceKey)
			}
		}

		// Sleep for a short interval before checking again.
		time.Sleep(time.Second)
	}
}
