package actions

import (
	"encoding/json"
	"log"
	"os/exec"
	natsconn "sd/nats"

	"github.com/nats-io/nats.go"
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
			log.Printf("Error unmarshaling JSON: %v\n", err)
			return
		}

		// Define the command.
		cmd := exec.Command("sh", "-c", msg.Command)

 		// Run the command.
		if err := cmd.Run(); err != nil {
			log.Printf("Error executing command: %v\n", err)
		}

	})
}