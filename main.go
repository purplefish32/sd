package main

import (
	"encoding/json"
	"log"
	"os"

	iconBuilderSubscribers "sd/iconbuilder/subscribers"
	"sd/streamdeck"
	"sd/streamdeck/xl"
	"sd/streamdeck/xl/publishers"
	streamdeckXlSubscribers "sd/streamdeck/xl/subscribers"
	"sd/streamdeck/xl/utils"

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

	// List all HID devices
	devices := hid.Enumerate(streamdeck.VendorID, xl.ProductID)

	if len(devices) == 0 {
		log.Fatalf("No Stream Deck found with VendorID %04x and ProductID %04x", streamdeck.VendorID, xl.ProductID)
	}

	// Open the first Stream Deck found
	device, err := devices[0].Open()
	defer device.Close()

	if err != nil {
		log.Fatalf("Failed to open Stream Deck: %v", err)
	}

	streamdeckXlSubscribers.SubscribeInitialize(nc, device)
	streamdeckXlSubscribers.SubscribeUpdate(nc, device)
	streamdeckXlSubscribers.SubscribeAction(nc, device)
	iconBuilderSubscribers.SubscribeIconBuilderCreateBuffer(nc)

	publishers.SdInitialize(nc)

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