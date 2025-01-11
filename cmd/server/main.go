package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

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

	log.Info().Msg("Watching Stream Decks")
	// Keep the main program running.
	select {}
}
