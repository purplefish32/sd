package keyboard

import (
	"encoding/json"
	"sd/natsconn"

	"github.com/go-vgo/robotgo"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
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
			log.Error().Err(err).Msg("Error unmarshaling JSON")
			return
		}

		// Perform the desired action.
		robotgo.TypeStr(msg.Text)
	})
}