package watchers

import (
	"time"

	"github.com/karalabe/hid"
	"github.com/rs/zerolog/log"
)

const (
	VendorIDElgato = 0x0fd9
)

func WatchStreamDecks(instanceID string, onConnect ConnectHandler, onDisconnect DisconnectHandler) error {
	// Track currently connected devices
	connectedDevices := make(map[string]bool)

	for {
		// Find all Stream Deck devices
		devices := hid.Enumerate(VendorIDElgato, 0)

		// Track current devices for this iteration
		currentDevices := make(map[string]bool)

		for _, device := range devices {
			deviceID := device.Serial

			currentDevices[deviceID] = true

			// If device wasn't previously connected, trigger connect handler
			if !connectedDevices[deviceID] {
				if err := onConnect(instanceID, deviceID, uint16(device.ProductID)); err != nil {
					log.Error().Err(err).
						Str("deviceID", deviceID).
						Msg("Failed to handle device connection")
				}
			}
		}

		// Check for disconnected devices
		for deviceID := range connectedDevices {
			if !currentDevices[deviceID] {
				if err := onDisconnect(instanceID, deviceID); err != nil {
					log.Error().Err(err).
						Str("deviceID", deviceID).
						Msg("Failed to handle device disconnection")
				}
				delete(connectedDevices, deviceID)
			}
		}

		// Update connected devices map
		connectedDevices = currentDevices

		time.Sleep(time.Second) // Poll interval
	}
}
