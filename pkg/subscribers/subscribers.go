package subscribers

import (
	"encoding/json"
	"sd/natsconn"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type UpdateMessageData struct {
	Key   string `json:"key"`
	Image string `json:"image"`
}

type UpdateMessage struct {
	Id      string            `json:"id"`
	Pattern string            `json:"pattern"`
	Data    UpdateMessageData `json:"data"`
}

func SubscribeToUpdateMessages() {
	nc, _ := natsconn.GetNATSConn()

	nc.Subscribe("sd.update", func(m *nats.Msg) {
		// Parse the JSON message
		var updateMessage UpdateMessage

		err := json.Unmarshal(m.Data, &updateMessage)

		if err != nil {
			log.Error().Err(err).Msg("Failed to parse JSON")
			return
		}
	})
}

// TODO update events:
// Update Image
// Update Settings
// Update Title
// Update State
// Update Current Profile
// Update Current Page
// Create Profile
// Create Page
// Delete Profile
// Delete Page
