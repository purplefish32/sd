package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"sd/pkg/core"
	"sd/pkg/env"
	"sd/pkg/instance"
	"sd/pkg/natsconn"
	"sd/pkg/plugins/browser"
	"sd/pkg/plugins/command"
	"sd/pkg/plugins/keyboard"
	"sd/pkg/streamdeck/xl"
	"sd/pkg/util"
	"sd/pkg/watchers"

	"github.com/karalabe/hid"
)

const (
	// Device Types
	DeviceTypeXL    = "xl"
	DeviceTypePlus  = "plus"
	DeviceTypePedal = "pedal"

	// USB IDs
	VendorIDElgato = 0x0fd9
	ProductIDXL    = 0x006c // Stream Deck XL
	ProductIDPlus  = 0x0084 // Stream Deck +
	ProductIDPedal = 0x0086 // Stream Deck Pedal
)

type DeviceInfo struct {
	Type      string    `json:"type"`       // xl, plus, pedal
	CreatedAt time.Time `json:"created_at"` // When the device was first seen
	UpdatedAt time.Time `json:"updated_at"` // Last time device was seen
	Status    string    `json:"status"`     // connected, disconnected
}

func DetermineDeviceType(productID uint16) string {
	switch productID {
	case ProductIDXL:
		return DeviceTypeXL
	case ProductIDPlus:
		return DeviceTypePlus
	case ProductIDPedal:
		return DeviceTypePedal
	default:
		return "unknown"
	}
}

func disconnectDevice(instanceID string, deviceID string, status string) error {
	_, kv := natsconn.GetNATSConn()
	if kv == nil {
		return fmt.Errorf("failed to get NATS KV store")
	}

	key := fmt.Sprintf("instances.%s.devices.%s", instanceID, deviceID)

	entry, err := kv.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get device info: %w", err)
	}

	var info DeviceInfo
	if entry != nil {
		if err := json.Unmarshal(entry.Value(), &info); err != nil {
			return fmt.Errorf("failed to unmarshal device info: %w", err)
		}
	}

	info.Status = status
	info.UpdatedAt = time.Now()

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	_, err = kv.Put(key, data)
	if err != nil {
		return fmt.Errorf("failed to store device info: %w", err)
	}

	log.Info().
		Str("instance", instanceID).
		Str("device", deviceID).
		Str("status", status).
		Msg("Updated device status")

	return nil
}

func connectDevice(instanceID string, deviceID string, productID uint16) error {
	_, kv := natsconn.GetNATSConn()
	if kv == nil {
		return fmt.Errorf("failed to get NATS KV store")
	}

	deviceType := DetermineDeviceType(productID)

	if deviceType == "unknown" {
		return fmt.Errorf("unknown device type for product ID: %x", productID)
	}

	// Replace the direct Open call with enumeration and open
	devices := hid.Enumerate(VendorIDElgato, productID)
	if len(devices) == 0 {
		return fmt.Errorf("no devices found with product ID: %x", productID)
	}

	device, err := devices[0].Open()
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}

	// Initialize the appropriate device type
	var initErr error
	switch deviceType {
	case DeviceTypeXL:
		xlDevice := xl.New(instanceID, device)
		initErr = xlDevice.Init()
		// case DeviceTypePlus:
		// 	plusDevice := plus.New(instanceID, device)
		// 	initErr = plusDevice.Init()
		// case DeviceTypePedal:
		// 	pedalDevice := pedal.New(instanceID, device)
		// 	initErr = pedalDevice.Init()
	}

	if initErr != nil {
		device.Close()
		return fmt.Errorf("failed to initialize %s device: %w", deviceType, initErr)
	}

	// Store device info in NATS KV
	key := fmt.Sprintf("instances.%s.devices.%s", instanceID, deviceID)
	info := DeviceInfo{
		Type:      deviceType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    "connected",
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	_, err = kv.Put(key, data)
	if err != nil {
		return fmt.Errorf("failed to store device info: %w", err)
	}

	log.Info().
		Str("instance", instanceID).
		Str("device", deviceID).
		Str("type", deviceType).
		Msg("Device connected and initialized")

	return nil
}

func main() {
	// Set global time format for logger.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Configure the global logger.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Str("app", "server").Logger()

	log.Info().Msg("Starting application")

	// Retrieve or create the instance UUID.
	instanceID := instance.GetOrCreateInstanceUUID()

	// Load environment variables
	root, err := util.GetProjectRoot()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get project root")
		return
	}

	env.LoadEnv(root + "/cmd/server/.env")

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

	// Start watching Stream Deck devices with connect/disconnect handlers
	go func() {
		err := watchers.WatchStreamDecks(
			instanceID,
			// Connected handler
			func(instanceID string, deviceID string, productID uint16) error {
				return connectDevice(instanceID, deviceID, productID)
			},
			// Disconnected handler
			func(instanceID string, deviceID string) error {
				return disconnectDevice(instanceID, deviceID, "disconnected")
			},
		)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to watch Stream Decks")
		}
	}()

	log.Info().Msg("Watching Stream Decks")
	// Keep the main program running.
	select {}
}
