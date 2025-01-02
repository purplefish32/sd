package plus

import (
	"encoding/json"
	"sd/pkg/actions"
	"sd/pkg/env"
	"sd/pkg/natsconn"
	"sd/pkg/pages"
	"sd/pkg/profiles"
	"sd/pkg/util"
	"strconv"
	"strings"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Plus struct {
	instanceID string
	device     *hid.Device
}

var ProductID uint16 = 0x006c

func New(instanceID string, device *hid.Device) Plus {
	return Plus{
		instanceID: instanceID,
		device:     device,
	}
}

func (plus Plus) Init() {
	log.Info().
		Str("device_serial", plus.device.Serial).
		Msg("Stream Deck XL Initialization")

	// Blank all keys.
	BlankAllKeys(plus.device)

	currentProfile := profiles.GetCurrentProfile(plus.instanceID, plus.device)

	// If no default profile exists, create one and set is as the default profile.
	if currentProfile == nil {
		log.Warn().Msg("Current profile not found creating one")

		// Create a new profile.
		profile, _ := profiles.CreateProfile(plus.instanceID, plus.device, "Default")

		log.Info().Str("profileId", profile.ID).Msg("Profile created")

		// Set the profile as the current profile.
		profiles.SetCurrentProfile(plus.instanceID, plus.device.Serial, profile.ID)
	}

	currentProfile = profiles.GetCurrentProfile(plus.instanceID, plus.device)

	log.Info().Interface("current_profile", currentProfile).Msg("Current profile")

	currentPage := pages.GetCurrentPage(plus.instanceID, plus.device, currentProfile.ID)

	// If no default page exists, create one and set is as the default page for the given profile.
	if currentPage == nil {
		log.Warn().Msg("Current page not found creating one")

		// Create a new page.
		page := pages.CreatePage(plus.instanceID, plus.device, currentProfile.ID)

		log.Info().Interface("page", page).Msg("Page created")

		// Set the page as the current page.
		pages.SetCurrentPage(plus.instanceID, plus.device, currentProfile.ID, page.ID)
	}

	// Buffer for outgoing events.
	buf := make([]byte, 512)

	// Get NATS connection an KV store.
	nc, kv := natsconn.GetNATSConn()

	go WatchKVForButtonImageBufferChanges(plus.instanceID, plus.device)
	go WatchForButtonChanges()

	// Listen for incoming device input.
	for {
		n, _ := plus.device.Read(buf)

		if n > 0 {
			pressedButtons := util.ParseEventBuffer(buf)

			// TODO implement long press.
			for _, buttonIndex := range pressedButtons {

				// Ignore button up event for now.
				if buttonIndex == 0 {
					log.Debug().Interface("device", plus.device).Int("button_index", buttonIndex).Msg("Button released")
					continue
				}

				log.Debug().Interface("device", plus.device).Int("button_index", buttonIndex).Msg("Button pressed")

				key := "instances." + plus.instanceID + ".devices." + plus.device.Serial + ".profiles." + currentProfile.ID + ".pages." + currentPage.ID + ".buttons." + strconv.Itoa(buttonIndex)

				log.Debug().Msg(key)

				// Get the associated data from the NATS KV Store.
				entry, err := nats.KeyValue.Get(kv, key)

				if err != nil {
					log.Warn().Err(err).Msg("Failed to get value from KV store")
					continue
				}

				// Unmarshal the JSON into the Payload struct
				var payload actions.ActionInstance

				if err := json.Unmarshal(entry.Value(), &payload); err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal JSON from KV store")
					return
				}

				log.Debug().Interface("payload", payload).Msg("NATS KV store data")

				// Use the `UUID` field as the topic
				if payload.UUID == "" {
					log.Error().Msg("Missing UUID field in JSON payload")
					return
				}

				// Publish Action Instance to NATS.
				nc.Publish(payload.UUID, entry.Value())
			}
		}
	}
}

func BlankKey(device *hid.Device, keyId int, buffer []byte) {
	// Update Key.
	util.SetKeyFromBufferPlus(device, keyId, buffer)
}

func BlankAllKeys(device *hid.Device) {
	var assetPath = env.Get("ASSET_PATH", "")
	var buffer, err = util.ConvertImageToBuffer(assetPath+"images/black.jpg", 250)

	if err != nil {
		log.Error().Err(err).Msg("Could not convert blank image to buffer")
	}

	for i := 1; i <= 8; i++ {
		BlankKey(device, i, buffer)
	}
}

func WatchForButtonChanges() {
	_, kv := natsconn.GetNATSConn()

	// Start watching the KV bucket for all button changes.
	watcher, err := kv.Watch("instances.*.devices.*.profiles.*.pages.*.buttons.*")
	defer watcher.Stop()

	if err != nil {
		log.Error().Err(err).Msg("Error creating watcher")
	}

	// Start the watch loop.
	for update := range watcher.Updates() {
		if update == nil {
			continue
		}

		// Parse JSON from update.Value().
		var actionInstance actions.ActionInstance

		err := json.Unmarshal(update.Value(), &actionInstance)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal JSON")
			continue
		}

		// TODO take into account multiple states ?
		buf, err := util.ConvertImageToBuffer(actionInstance.States[0].ImagePath, 250)

		if err != nil {
			log.Error().Err(err).Msg("Buffer error")
		}

		// Put the serialized data into the KV store.
		if _, err := kv.Put(string(update.Key())+".buffer", buf); err != nil {
			log.Error().Err(err).Msg("Error")
		}
	}
}
func WatchKVForButtonImageBufferChanges(instanceId string, device *hid.Device) {
	// Add contextual information to the logger for this function
	log := log.With().
		Str("instanceId", instanceId).
		Str("deviceSerial", device.Serial).
		Logger()

	_, kv := natsconn.GetNATSConn()

	currentProfile := profiles.GetCurrentProfile(instanceId, device)
	currentPage := pages.GetCurrentPage(instanceId, device, currentProfile.ID)

	// Start watching the KV bucket for updates for a specific profile and page.
	watcher, err := kv.Watch("instances." + instanceId + ".devices." + device.Serial + ".profiles." + currentProfile.ID + ".pages." + currentPage.ID + ".buttons.*.buffer")
	defer watcher.Stop()

	if err != nil {
		log.Error().Err(err).Msg("Error creating watcher")
	}

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

			// Split the string by the delimiter ".".
			segments := strings.Split(update.Key(), ".")

			// Get the last segment.
			sdKeyId := segments[len(segments)-2]

			// Convert to an int.
			id, err := strconv.Atoi(sdKeyId)

			if err != nil {
				// ... handle error.
				panic(err)
			}

			// Update Key.
			util.SetKeyFromBufferPlus(device, id, update.Value())
		case nats.KeyValueDelete:
			log.Info().Str("key", update.Key()).Msg("Key deleted")
		default:
			log.Info().Str("key", update.Key()).Msg("Unknown operation")
		}
	}
}
