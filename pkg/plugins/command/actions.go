package command

import (
	"encoding/json"
	"os/exec"
	"sd/pkg/env"
	"sd/pkg/natsconn"
	"sd/pkg/types"
	"sd/pkg/util"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Settings struct {
	Command string `json:"command"`
}

// Subscribe sets up the NATS subscription for this plugin.
func OpenSubscriber() {
	root, err := util.GetProjectRoot()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get project root")
		return
	}

	env.LoadEnv(root + "/pkg/plugins/.env")

	nc, _ := natsconn.GetNATSConn()

	if _, err := nc.Subscribe("sd.plugin.command.exec", func(m *nats.Msg) {
		var actionInstance types.ActionInstance

		// Parse the incoming message
		if err := json.Unmarshal(m.Data, &actionInstance); err != nil {
			log.Error().Err(err).Msg("Error unmarshaling JSON")
			return
		}

		// Convert actionInstance.Settings to Settings
		settingsMap, ok := actionInstance.Settings.(map[string]any)

		if !ok {
			log.Error().Msg("Settings is not a valid object")
			return
		}

		settingsBytes, err := json.Marshal(settingsMap)

		if err != nil {
			log.Error().Err(err).Msg("Error marshaling settings to JSON")
			return
		}

		var settings Settings

		if err := json.Unmarshal(settingsBytes, &settings); err != nil {
			log.Error().Err(err).Msg("Error unmarshaling settings to Settings")
			return
		}

		// Validate URL
		if settings.Command == "" {
			log.Error().Msg("Command is empty")
			return
		}

		cmd := exec.Command("sh", "-c", settings.Command)

		// Detach the process
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true, // Create a new session
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			log.Error().Err(err).Msg("Cannot start command")
			return
		}

	}); err != nil {
		log.Fatal().Err(err).Msg("Failed to subscribe to sd.plugin.command.exec")
	}
}
