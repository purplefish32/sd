package subscribers

import (
	"encoding/json"
	"log"

	"sd/streamdeck/xl/utils"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

type UpdateMessageData struct {
	Key int `json:"key"`
	Image string `json:"image"`
}

type UpdateMessage struct {
	Id string `json:"id"`
	Pattern string `json:"pattern"`
	Data UpdateMessageData `json:"data"`
}

func SubscribeSdUpdate(nc *nats.Conn, device *hid.Device) {
	nc.Subscribe("sd.update", func(m *nats.Msg) {

		// Parse the JSON message
		var event UpdateMessage

		err := json.Unmarshal(m.Data, &event)

		if err != nil {
			log.Printf("Failed to parse JSON message: %v", err)
			return
		}

		utils.SetKey(device, event.Data.Key-1, event.Data.Image)
	})
}