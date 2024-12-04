package xl

import (
	"encoding/json"
	natsconn "sd/nats"
	"sd/streamdeck/xl/buttons"
	"sd/streamdeck/xl/pages"
	"sd/streamdeck/xl/profiles"
	"sd/streamdeck/xl/utils"
	"strconv"
	"strings"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

const ProductID = 0x006c

type ButtonEvent struct {
	Id int `json:"id"`
	Type string `json:"type"`
	Device string `json:"device"`
	Model string `json:"model"`
	InstanceID string `json:"instanceId"`
	Key string `json:"key"`
}

type UpdateMessageData struct {
	Key string `json:"key"`
	Image string `json:"image"`
}

type UpdateMessage struct {
	Id string `json:"id"`
	Pattern string `json:"pattern"`
	Data UpdateMessageData `json:"data"`
}

func Subscribe(instanceId string, device *hid.Device) {
	nc, _ := natsconn.GetNATSConn()

	log.Info().
		Str("instance_id", instanceId).
		Str("device_serial", device.Serial).
		Msg("Subscribing to sd.update events")

	nc.Subscribe("sd.update", func(m *nats.Msg) {
		log.Info().Str("device_serial", device.Serial).Str("message", string(m.Data)).Msg("Received a message on sd.update channel", )

		// Parse the JSON message
		var updateMessage UpdateMessage

		err := json.Unmarshal(m.Data, &updateMessage)

		if err != nil {
			log.Error().Err(err).Msg("Failed to parse JSON")
			return
		}

		// TODO update events:
		// Update Image
		// Update Settings
		// Update Title
		// Update State
		// Update Current Profile
		// Update Current Page
		// Create Profile
		// Create Page
		// Delete Profile
		// Delete Page

		// TODO get button, update button
		//log.Printf("HERE IS MY KEY: %+v", updateMessage.Data.Key)
		//button, err := buttons.GetButton(kv, updateMessage.Data.Key)
		//log.Printf("HERE IS MY BUTTON: %+v", button)

		// TODO
		buttons.UpdateButton("instances." + instanceId + ".devices." + device.Serial + ".profiles.ec3217e3-8713-4b86-8ec3-1c143877a72b.pages.c519c254-00e9-4000-8277-35f7be8af772.buttons.0")


		//utils.SetKey(device, event.Data.Key-1, event.Data.Image)

		// SetKVIconFromImage(kv, instanceID, device, event.Data.Key-1, event.Data.Image)
	})
}

func Init(instanceID string, device *hid.Device) {
	nc, _ := natsconn.GetNATSConn()

	log.Info().
		Str("device_serial", device.Serial).
		Msg("Stream Deck XL Initialization")

	Subscribe(instanceID, device);

	currentProfile, _ := profiles.GetCurrentProfile(instanceID, device);

	// If no default profile exists, create one and set is as the default profile.
	if currentProfile == nil {
		// Create a new profile.
		profileId, _ := profiles.CreateProfile(instanceID, device, "Default");

		// Set the profile as the current profile.
		profiles.SetCurrentProfile(instanceID, device, profileId)
	}

	go WatchKVForButtonImageBufferChanges(instanceID, device)


	// TEMP // TODO

	if currentProfile != nil {
		currentPage, _ := pages.GetCurrentPage(instanceID, device, currentProfile.ID)
		var updateMessage = UpdateMessage{
			Id: "",
			Pattern: "",
			Data: UpdateMessageData{
				Key: "instances." + instanceID + ".devices." + device.Serial + ".profiles." + currentProfile.ID + ".pages." + currentPage.ID + ".buttons.1",
				Image: "./assets/images/red.jpg",
			},
		}

		// Marshal the event struct to JSON.
		payload, _ := json.Marshal(updateMessage)

		nc.Publish("sd.update", payload)
	}
	// END TEMP

	// Buffer for outgoing events.
	buf := make([]byte, 512)

	for {
		n, err := device.Read(buf)

		if err != nil {
			log.Error().Err(err).Msg("Error reading from Stream Deck")
			continue
		}

		if n > 0 {
			pressedButtons := utils.ParseEventBuffer(buf)

			profile, _ := profiles.GetCurrentProfile(instanceID, device)
			page, _ := pages.GetCurrentPage(instanceID, device, profile.ID)

			for _, buttonIndex := range pressedButtons {
				// Create a new buttonEvent struct for each pressed button.
				event := ButtonEvent{
					Id: buttonIndex,
					Type: "key",
					Device: device.Serial,
					Model: "XL",
					InstanceID: instanceID,
					Key: "instances." + instanceID + ".devices." + device.Serial + ".profiles." + profile.ID + ".pages." + page.ID + ".buttons." +  strconv.Itoa(buttonIndex) ,
				}

				// Marshal the event struct to JSON.
				payload, _ := json.Marshal(event)

				// Publish the JSON payload to the NATS topic.
				nc.Publish("sd.event", payload)

				log.Info().Msg("Published sd.event")
			}
		}
	}
}

// func SetKeyBufferFromImagePath(kv nats.KeyValue, instanceId string, device *hid.Device, profileId string, pageId string, buttonId string, imagePath string) {
// 	buf, _ := bimg.Read(imagePath)

// 	kv.Put("instances." + instanceId + ".devices." + device.Serial + ".profiles." + profileId + ".page." + pageId + ".buttons." + buttonId, buf)
// }

// func SetKVIconFromImage(kv nats.KeyValue, instanceID string, device *hid.Device, buttonId int, imagePath string) {
// 	buf, _ := bimg.Read(imagePath)

// 	kv.Put("instances." + instanceID + ".devices." + device.Serial + ".profiles.xxx.page.xxx.buttons." + fmt.Sprintf("%v", buttonId), buf)
// }

// WatchKV watches for changes in the given KeyValue store
func WatchKVForButtonImageBufferChanges(instanceId string, device *hid.Device) {
	_, kv := natsconn.GetNATSConn()

	log.Info().
		Str("instance_id", instanceId).
		Str("device_serial", device.Serial).
		Msg("Starting KV Watcher")

	currentProfile, _ := profiles.GetCurrentProfile(instanceId, device)
	currentPage, _ := pages.GetCurrentPage(instanceId, device, currentProfile.ID)

	// Start watching the KV bucket for all updates.
	watcher, err := kv.Watch("instances." + instanceId + ".devices." + device.Serial + ".profiles." + currentProfile.ID + ".pages." + currentPage.ID + ".buttons.*.buffer" )

	if err != nil {
		log.Error().Err(err).Msg("Error creating watcher")
	}

	defer watcher.Stop()

	// Flag to track when all initial values have been processed.
	initialValuesProcessed := false

	// Start the watch loop.
	for update := range watcher.Updates() {
		// If the update is nil, it means all initial values have been received.
		if update == nil {
			if !initialValuesProcessed {
				log.Info().Msg("All initial values have been processed. Waiting for updates")
				initialValuesProcessed = true
			}
			// Continue listening for future updates, so don't break here.
			continue
		}

		// Process the update.
		switch update.Operation() {
			case nats.KeyValuePut:
				log.Info().Str("key", update.Key()).Msg("Key added/updated")
				// Get Stream Deck key id from the kv key.

				// Split the string by the delimiter "."
				segments := strings.Split(update.Key(), ".")

				// Get the last segment
				sdKeyId := segments[len(segments)-2]

				// Concert to int
				id, err := strconv.Atoi(sdKeyId)

				if err != nil {
					// ... handle error
					panic(err)
				}

				// Update Key.
				utils.SetKeyFromBuffer(device, id, update.Value())
			case nats.KeyValueDelete:
				log.Info().Str("key", update.Key()).Msg("Key deleted")
			default:
				log.Info().Str("key", update.Key()).Msg("Unknown operation")
		}
	}
}