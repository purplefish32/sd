package actions

import (
	"encoding/json"
	"log"

	natsconn "sd/nats"

	"github.com/go-vgo/robotgo"
	"github.com/nats-io/nats.go"
)

var msg struct {
	Text   string `json:"text"`
}

// Subscribe sets up the NATS subscription for this plugin.
func SubscribeActionType(pluginNamespace string) {
	nc, _ := natsconn.GetNATSConn()

	nc.Subscribe(pluginNamespace + ".type", func(m *nats.Msg) {

		// Parse the incoming message.
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			log.Printf("Error unmarshaling JSON: %v\n", err)
			return
		}

		// Perform the desired action.
		robotgo.TypeStr(msg.Text)
	})
}