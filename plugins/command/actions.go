package command

import (
	"encoding/json"
	"os/exec"
	natsconn "sd/nats"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

var msg struct {
	Command string `json:"command"`
}

// Subscribe sets up the NATS subscription for this plugin.
func SubscribeActionExec(pluginNamespace string) {
	nc, _ := natsconn.GetNATSConn()

	nc.Subscribe(pluginNamespace + ".exec", func(m *nats.Msg) {

		// Parse the incoming message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			log.Error().Err(err).Msg("Error unmarshaling JSON")
			return
		}

		// Define the command.
		cmd := exec.Command("sh", "-c", msg.Command)

 		// Run the command.
		if err := cmd.Run(); err != nil {
			log.Error().Err(err).Msg("Can not run command")
		}

	})
}