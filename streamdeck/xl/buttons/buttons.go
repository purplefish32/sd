package buttons

import (
	"encoding/json"
	"log"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

type ButtonMetadata struct {
}

type Button struct {
	Plugin string `json:"plugin"`
	Action string `json:"action"`
	Image string `json:"image"`
	MetaData ButtonMetadata `json:"metadata"`
}

// CreateButton
func CreateButton(kv nats.KeyValue, instanceId string, device *hid.Device, profileId string, pageId string, buttonId string) (err error) {
	// Define the key for the current button
	key := "instances." + instanceId + ".devices." + device.Serial + ".profiles." + profileId + ".pages." + pageId + ".buttons." + buttonId

	// Define a new Button.
	button := Button{
		Plugin: "",
		Action: "",
		Image: "./assets/images/black.jpg",
		MetaData: ButtonMetadata{},
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(button)

	if err != nil {
		log.Printf("Failed to serialize button data: %v", err)
		return err
	}

	// Put the serialized data into the KV store
	_, err = kv.Create(key, data)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Button key already exists: %s", key)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", key, err)
		}
		return err
	}

	log.Printf("Button %v created successfully: %+v", buttonId, button)

	return nil
}

// TODO
// UpdateButton
// ResetButton ?
// MoveButton ?