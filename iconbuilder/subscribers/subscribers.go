package subscribers

import (
	"encoding/json"
	"log"

	"sd/iconbuilder"
	"sd/iconbuilder/utils"

	"github.com/nats-io/nats.go"
)

func SubscribeIconBuilderCreateBuffer(nc *nats.Conn) {
	// Subscribe to the subject
	subject := "sd.iconbuilder.buffer.create"
	nc.Subscribe(subject, func(m *nats.Msg) {
		var req iconbuilder.IconRequest
		if err := json.Unmarshal(m.Data, &req); err != nil {
			log.Printf("Invalid request: %v", err)
			return
		}

		icon, _ := utils.CreateIconBuffer(req)

		nc.Publish(m.Reply, icon)
	})

	log.Println("Listening for icon creation requests on", subject)
}