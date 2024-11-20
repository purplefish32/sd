package xl

import (
	"encoding/json"
	"fmt"
	"log"
	streamdeckXlSubscribers "sd/streamdeck/xl/subscribers"
	"sd/streamdeck/xl/utils"
	"sync"

	"github.com/h2non/bimg"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

const ProductID = 0x006c


type buttonEvent struct {
	Id int `json:"id"`
	Type string `json:"type"`
	Serial string `json:"serial"`
	InstanceID string `json:"instanceId"`
}

func Initialize(nc *nats.Conn, instanceID string, device *hid.Device, kv nats.KeyValue) {
	log.Println("Stream Deck XL Initialization")

	buf, err := bimg.Read("./assets/images/black.jpg")

	if err != nil {
		log.Fatal("Error reading image:", err)
	}

	// Create all missing keys in Jetstream KV.
	var wg sync.WaitGroup // To wait for all goroutines to finish.

	for i := 0; i < 32; i++ {
		// Increment the WaitGroup counter.
		wg.Add(1)

		// Generate the key for the current iteration.
		key := instanceID + "." + device.Serial + "." + fmt.Sprintf("%v", i)

		// Launch a goroutine for each iteration.
		go func(i int, key string) {
			defer wg.Done() // Decrement the counter when the goroutine finishes.

			// Create the value (you can replace `buf` with the actual data you want to store).
			_, err := kv.Create(key, []byte(buf))
			if err != nil {
				log.Printf("Error creating key %s: %v", key, err)
				return
			}

			// Get the key-value entry.
			entry, err := kv.Get(key)
			if err != nil {
				log.Printf("Error getting key %s: %v", key, err)
				return
			}

			// Use the buffer (or call your utility function).
			utils.SetKeyFromBuffer(device, i, entry.Value())
		}(i, key)
	}

	// Wait for all goroutines to complete.
	wg.Wait()

	log.Println("Rendering all icons completed")

	// Listen for key updates.
	streamdeckXlSubscribers.UpdateKey(nc, device)

	// Buffer for outgoing events.
	buf = make([]byte, 512)

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

					// Create a new buttonEvent struct for each pressed button.
					event := buttonEvent{
						Id: buttonIndex,
						Type: "XL",
						Serial: device.DeviceInfo.Serial,
						InstanceID: instanceID,
					}

					// Marshal the event struct to JSON.
					eventJSON, _ := json.Marshal(event)

					// Publish the JSON payload to the NATS topic.
					nc.Publish("sd.event", eventJSON)
				}
			}
		}
	}
}