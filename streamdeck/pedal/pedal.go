package streamdeck

import (
	"encoding/json"
	natsconn "sd/nats"
	"sd/streamdeck/xl/utils"

	"github.com/karalabe/hid"
	"github.com/rs/zerolog/log"
)

const ProductID = 0x0086

type buttonEvent struct {
	Id int `json:"id"`
	Type string `json:"type"`
	Serial string `json:"serial"`
	InstanceID string `json:"instanceId"`
}

func Init(instanceID string, device *hid.Device) {
	log.Info().Str("device_serial", device.Serial).Msg("Stream Deck Pedal Initialization")
	nc, _ := natsconn.GetNATSConn()

	// Buffer for outgoing events.
	buf := make([]byte, 512)

	for {
		n, err := device.Read(buf)

		if err != nil {
			log.Error().Str("device_serial", device.Serial).Err(err).Msg("Unable to read buffer")
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