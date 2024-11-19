package subscribers

import (
	"encoding/json"
	"log"

	"sd/iconbuilder"
	"sd/iconbuilder/utils"

	"github.com/nats-io/nats.go"
)

func SubscribeIconBuilderCreate(nc *nats.Conn) {
	// Subscribe to the subject
	subject := "sd.iconbuilder.create"
	nc.Subscribe(subject, func(m *nats.Msg) {
		var req iconbuilder.IconRequest
		if err := json.Unmarshal(m.Data, &req); err != nil {
			log.Printf("Invalid request: %v", err)
			return
		}

		icon, err := utils.Create(req)
		response := iconbuilder.IconResponse{Success: err == nil}
		if err != nil {
			response.Message = err.Error()
		} else {
			response.Message = "Icon created successfully"
			response.Data = icon
		}

		respData, _ := json.Marshal(response)
		nc.Publish(m.Reply, respData)
	})

	log.Println("Listening for icon creation requests on", subject)
}