package command

import (
	"encoding/json"
	"log"
	"os/exec"

	"github.com/nats-io/nats.go"
)

// CommandPlugin represents the command plugin
type CommandPlugin struct{}

// Name returns the name of the plugin
func (c *CommandPlugin) Name() string {
	return "command"
}

// Subscribe sets up the NATS subscription for this plugin
func (c *CommandPlugin) Subscribe(nc *nats.Conn) error {
	_, err := nc.Subscribe("sd.plugin.command", func(m *nats.Msg) {
		log.Println("Command Plugin received NATS message:", string(m.Data))

		var msg struct {
			Data struct {
				Action  string `json:"action"`
				Command string `json:"command"`
			} `json:"data"`
		}

		// Parse the incoming message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			log.Printf("Error unmarshaling JSON: %v\n", err)
			return
		}

		// Perform the desired action
		if msg.Data.Action == "exec" {
			log.Println("Executing command:", msg.Data.Command)
			cmd := exec.Command("sh", "-c", msg.Data.Command) // Run the command
			if err := cmd.Run(); err != nil {
				log.Printf("Error executing command: %v\n", err)
			}
		}
	})
	return err
}