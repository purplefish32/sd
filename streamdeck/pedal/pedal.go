package streamdeck

import (
	"encoding/json"
	"log"
	"sd/streamdeck/xl/utils"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

const ProductID = 0x0086

type buttonEvent struct {
	Id int `json:"id"`
	Type string `json:"type"`
	Serial string `json:"serial"`
	InstanceID string `json:"instanceId"`
}

func Initialize(nc *nats.Conn, instanceID string, device *hid.Device, kv nats.KeyValue) {
	log.Println("Stream Deck Pedal Initialization")

	// Buffer for outgoing events.
	buf := make([]byte, 512)

	for {
		n, err := device.Read(buf)

		if err != nil {
			log.Printf("Error reading from Stream Deck Pedal: %v", err)
			continue
		}


		if n > 0 {
			pressedButtons := utils.ParseEventBuffer(buf)

			if len(pressedButtons) > 0 {
				for _, buttonIndex := range pressedButtons {

					// Create a new buttonEvent struct for each pressed button
					event := buttonEvent{
						Id: buttonIndex,
						Type: "Pedal",
						Serial: device.DeviceInfo.Serial,
						InstanceID: instanceID,
					}

					// Marshal the event struct to JSON
					eventJSON, _ := json.Marshal(event)

					// Publish the JSON payload to the NATS topic
					nc.Publish("sd.event", eventJSON)
				}
			}
		}
	}
}