package xl

import (
	"encoding/json"
	"log"
	"sd/streamdeck/xl/profiles"
	"sd/streamdeck/xl/utils"
	"strconv"
	"strings"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

const ProductID = 0x006c

type ButtonEvent struct {
	Id int `json:"id"`
	Type string `json:"type"`
	Serial string `json:"serial"`
	InstanceID string `json:"instanceId"`
}

type UpdateMessageData struct {
	Key int `json:"key"`
	Image string `json:"image"`
}

type UpdateMessage struct {
	Id string `json:"id"`
	Pattern string `json:"pattern"`
	Data UpdateMessageData `json:"data"`
}

func Subscribe(nc *nats.Conn, kv nats.KeyValue, instanceID string, device *hid.Device) {
	log.Printf("Subscribing to sd.update events for instance: %+v device: %+v", instanceID, device.Serial)

	nc.Subscribe("sd.update", func(m *nats.Msg) {
		log.Printf("Received a message on sd.update events for device: %+v", device.Serial)

		// Parse the JSON message
		var event UpdateMessage

		err := json.Unmarshal(m.Data, &event)

		if err != nil {
			log.Printf("Failed to parse JSON message: %v", err)
			return
		}

		//utils.SetKey(device, event.Data.Key-1, event.Data.Image)

		// SetKVIconFromImage(kv, instanceID, device, event.Data.Key-1, event.Data.Image)
	})
}

func Initialize(nc *nats.Conn, instanceID string, device *hid.Device, kv nats.KeyValue) {
	log.Printf("Stream Deck XL Initialization: %+v", device.Serial)

	Subscribe(nc, kv, instanceID, device);
	
	currentProfile, _ := profiles.GetCurrentProfile(kv, instanceID, device);

	// If no default profile exists, create one and set is as the default profile.
	if currentProfile == nil {
		// Create a new profile.
		profileId, _ := profiles.CreateProfile(kv, instanceID, device, "Default");

		// Set the profile as the current profile.
		profiles.SetCurrentProfile(kv, instanceID, device, profileId)
	}

	//buf, err := bimg.Read("./assets/images/black.jpg")

	// if err != nil {
	// 	log.Fatal("Error reading image:", err)
	// }

	log.Println("Rendering all icons completed")

	go WatchKV(kv, instanceID, device)

	// Buffer for outgoing events.
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

					// Create a new buttonEvent struct for each pressed button.
					event := ButtonEvent{
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

// func SetKVIconFromImage(kv nats.KeyValue, instanceID string, device *hid.Device, buttonId int, imagePath string) {
// 	buf, _ := bimg.Read(imagePath)

// 	kv.Put("instances." + instanceID + ".devices." + device.Serial + ".profiles.xxx.page.xxx.buttons." + fmt.Sprintf("%v", buttonId), buf)
// }

// WatchKV watches for changes in the given KeyValue store
func WatchKV(kv nats.KeyValue, instanceID string, device *hid.Device) {
	log.Printf("Starting KV Watcher")

	// Start watching the KV bucket for all updates.
	watcher, err := kv.Watch("instances." + instanceID + ".devices." + device.Serial + ".buttons.>", )

	if err != nil {
		log.Fatalf("Error creating watcher: %v", err)
	}
	defer watcher.Stop()

	// Flag to track when all initial values have been processed.
	initialValuesProcessed := false

	// Start the watch loop.
	for update := range watcher.Updates() {
		// If the update is nil, it means all initial values have been received.
		if update == nil {
			if !initialValuesProcessed {
				log.Println("All initial values have been processed. Waiting for updates.")
				initialValuesProcessed = true
			}
			// Continue listening for future updates, so don't break here.
			continue
		}

		// Process the update.
		switch update.Operation() {
			case nats.KeyValuePut:
				log.Printf("Key added/updated: %s", update.Key())
				// Get Stream Deck key id from the kv key.

				// Split the string by the delimiter "."
				segments := strings.Split(update.Key(), ".")

				// Get the last segment
				sdKeyId := segments[len(segments)-1]

				// Concert to int
				id, err := strconv.Atoi(sdKeyId)
				if err != nil {
					// ... handle error
					panic(err)
				}

				// Update Key.
				utils.SetKeyFromBuffer(device, id, update.Value())
			case nats.KeyValueDelete:
				log.Printf("Key deleted: %s", update.Key())
			default:
				log.Printf("Unknown operation on key: %s", update.Key())
		}
	}
}