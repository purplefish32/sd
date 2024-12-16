package publishers

import (
	"encoding/json"
	"log"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

type InitializationEvent struct {}

var event = InitializationEvent{}

func Initialize(nc *nats.Conn, device *hid.Device) {
	// Marshal the event struct to JSON
	eventJSON, err := json.Marshal(event)

	if(err != nil) {
		log.Fatalf("Could not publish Stream Deck XL initialization to NATS")
	}

	// Simple Publisher
	nc.Publish("sd.initialize", []byte(eventJSON))
}

