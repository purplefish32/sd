package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"sd/pkg/core"
	"sd/pkg/env"
	"sd/pkg/natsconn"
	"sd/pkg/plugins/browser"
	"sd/pkg/plugins/command"
	"sd/pkg/plugins/keyboard"
	"sd/pkg/store"
	"sd/pkg/streamdeck"
	"sd/pkg/types"
	"sd/pkg/util"
	"sd/pkg/watchers"
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

	var device types.Device

	if entry != nil {
		if err := json.Unmarshal(entry.Value(), &device); err != nil {
			return fmt.Errorf("failed to unmarshal device info: %w", err)
		}
	}

	device.Status = status

	data, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %w", err)
	}

	_, err = kv.Put(key, data)
	if err != nil {
		return fmt.Errorf("failed to store device: %w", err)
	}

	streamdeck.RemoveDevice(deviceID)

	return nil
}

func connectDevice(instanceID string, deviceID string, productID uint16) error {
	_, kv := natsconn.GetNATSConn()
	if kv == nil {
		return fmt.Errorf("failed to get NATS KV store")
	}

	key := fmt.Sprintf("instances.%s.devices.%s", instanceID, deviceID)

	// Try to get existing device first
	entry, _ := kv.Get(key)
	var device types.Device
	if entry != nil {
		if err := json.Unmarshal(entry.Value(), &device); err != nil {
			log.Warn().Err(err).Msg("Failed to unmarshal existing device")
		}
	}

	// Update device fields while preserving others
	device.ID = deviceID
	device.Instance = instanceID
	device.Type = DetermineDeviceType(productID)
	device.Status = "connected"

	data, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	_, err = kv.Put(key, data)
	if err != nil {
		return fmt.Errorf("failed to store device info: %w", err)
	}

	streamdeck.New(instanceID, deviceID, productID)
	return nil
}

func createLockFile() error {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create .streamdeck directory if it doesn't exist
	sdDir := filepath.Join(home, ".streamdeck")
	if err := os.MkdirAll(sdDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Try to create and lock the file
	lockFile := filepath.Join(sdDir, "server.lock")
	file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to get an exclusive lock
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		return fmt.Errorf("another instance is already running")
	}

	// Keep file open - lock will be released when process exits
	return nil
}

func main() {
	// Check for existing instance first
	if err := createLockFile(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Set global time format for logger.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Configure the global logger.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Str("app", "server").Logger()

	log.Info().Msg("Starting application")

	// Retrieve or create the instance UUID.
	instanceID := store.GetOrCreateInstanceUUID()

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
