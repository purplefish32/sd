package main

import (
	"encoding/json"
	"log"
	"os"

	xl "sd/stream_decks/xl/publishers"
	"sd/stream_decks/xl/subscribers"
	"sd/stream_decks/xl/utils"

	"github.com/joho/godotenv"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

type ButtonEvent struct {
	Id int `json:"id"`
}



func main() {
	// Load the .env file
	err := godotenv.Load()
	
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get NATS server address from the .env file
	natsServer := os.Getenv("NATS_SERVER")
	
	if natsServer == "" {
		log.Fatal("NATS_SERVER is not set in the .env file")
	}

	// Connect to NATS
	nc, err := nats.Connect(natsServer) // TODO This should be an array.
	
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}

    defer nc.Close()

	// Vendor and Product IDs for Stream Deck (adjust based on your model)
	const vendorID = 0x0fd9
	const productID = 0x006c

	// List all HID devices
	devices := hid.Enumerate(vendorID, productID)

	if len(devices) == 0 {
		log.Fatalf("No Stream Deck found with Vendor ID %04x and Product ID %04x", vendorID, productID)
	}

	// Open the first Stream Deck found
	device, err := devices[0].Open()
	defer device.Close()

	if err != nil {
		log.Fatalf("Failed to open Stream Deck: %v", err)
	}

	subscribers.SubscribeSdInitialize(nc, device)
	subscribers.SubscribeSdUpdate(nc, device)
	subscribers.SubscribeSdAction(nc, device)

	xl.PublishInitialization(nc)

	buf := make([]byte, 512)

	for {
		n, err := device.Read(buf)
		
		if err != nil {
			log.Printf("Error reading from Stream Deck: %v", err)
			continue
		}

		if n > 0 {
			pressedButtons := utils.ParseEventBuffer(buf)
			if len(pressedButtons) > 0 {
				for _, buttonIndex := range pressedButtons {
					// Create a new ButtonEvent struct for each pressed button
					event := ButtonEvent{
						Id: buttonIndex,
					}

					// Marshal the event struct to JSON
					eventJSON, err := json.Marshal(event)
					if err != nil {
						log.Printf("Error marshalling event to JSON: %v", err)
						continue
					}

					// Publish the JSON payload to the NATS topic
					err = nc.Publish("sd.event", eventJSON)
					if err != nil {
						log.Printf("Error publishing to NATS: %v", err)
					}
				}
			}
		}
	}
}