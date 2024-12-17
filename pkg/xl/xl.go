package xl

import (
	"encoding/json"
	"sd/pkg/actions"
	"sd/pkg/natsconn"
	"sd/pkg/pages"
	"sd/pkg/profiles"
	"sd/pkg/util"
	"sd/pkg/watchers"
	"strconv"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type XL struct {
	instanceID string
	device     *hid.Device
}

func New(instanceID string, device *hid.Device) XL {
	return XL{
		instanceID: instanceID,
		device:     device,
	}
}

func (xl XL) Init() {
	log.Info().
		Str("device_serial", xl.device.Serial).
		Msg("Stream Deck XL Initialization")

	currentProfile := profiles.GetCurrentProfile(xl.instanceID, xl.device)

	// If no default profile exists, create one and set is as the default profile.
	if currentProfile == nil {
		log.Warn().Msg("Current profile not found creating one")

		// Create a new profile.
		profile, _ := profiles.CreateProfile(xl.instanceID, xl.device, "Default")

		log.Info().Str("profileId", profile.ID).Msg("Profile created")

		// Set the profile as the current profile.
		profiles.SetCurrentProfile(xl.instanceID, xl.device, profile.ID)
	}

	currentProfile = profiles.GetCurrentProfile(xl.instanceID, xl.device)

	log.Info().Interface("current_profile", currentProfile).Msg("Current profile")

	currentPage := pages.GetCurrentPage(xl.instanceID, xl.device, currentProfile.ID)

	// If no default page exists, create one and set is as the default page for the given profile.
	if currentPage == nil {
		log.Warn().Msg("Current page not found creating one")

		// Create a new page.
		page := pages.CreatePage(xl.instanceID, xl.device, currentProfile.ID)

		log.Info().Interface("page", page).Msg("Page created")

		// Set the page as the current page.
		pages.SetCurrentPage(xl.instanceID, xl.device, currentProfile.ID, page.ID)
	}

	// Buffer for outgoing events.
	buf := make([]byte, 512)

	// Get NATS connection an KV store.
	nc, kv := natsconn.GetNATSConn()

	go watchers.WatchKVForButtonImageBufferChanges(xl.instanceID, xl.device)

	// Listen for incoming device input.
	for {
		n, err := xl.device.Read(buf)

		if err != nil {
			log.Error().Err(err).Msg("Error reading from Stream Deck")
			continue
		}

		if n > 0 {
			pressedButtons := util.ParseEventBuffer(buf)

			// TODO implement long press.
			for _, buttonIndex := range pressedButtons {

				// Ignore button up event for now.
				if buttonIndex == 0 {
					log.Debug().Interface("device", xl.device).Int("button_index", buttonIndex).Msg("Button released")
					continue
				}

				log.Debug().Interface("device", xl.device).Int("button_index", buttonIndex).Msg("Button pressed")

				key := "instances." + xl.instanceID + ".devices." + xl.device.Serial + ".profiles." + currentProfile.ID + ".pages." + currentPage.ID + ".buttons." + strconv.Itoa(buttonIndex)

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
