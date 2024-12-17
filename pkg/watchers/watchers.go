package watchers

import (
	"encoding/json"
	"sd/pkg/actions"
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
		buf := util.ConvertImageToBuffer(actionInstance.States[0].ImagePath)

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
			util.SetKeyFromBuffer(device, id, update.Value())
		case nats.KeyValueDelete:
			log.Info().Str("key", update.Key()).Msg("Key deleted")
		default:
			log.Info().Str("key", update.Key()).Msg("Unknown operation")
		}
	}
}
