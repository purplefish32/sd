package pedal

import (
	"encoding/json"
	"sd/pkg/actions"
	"sd/pkg/natsconn"
	"sd/pkg/profiles"
	"sd/pkg/util"
	"strconv"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Pedal struct {
	instanceID string
	device     *hid.Device
}

func New(instanceID string, device *hid.Device) Pedal {
	return Pedal{
		instanceID: instanceID,
		device:     device,
	}
}

func (pedal Pedal) Init() {
	log.Info().
		Str("device_serial", pedal.device.Serial).
		Msg("Stream Deck XL Initialization")

	currentProfile := profiles.GetCurrentProfile(pedal.instanceID, pedal.device)

	// If no default profile exists, create one and set is as the default profile.
	if currentProfile == nil {
		log.Warn().Msg("Current profile not found creating one")

		// Create a new profile.
		profile, _ := profiles.CreateProfile(pedal.instanceID, pedal.device, "Default")

		log.Info().Str("profileId", profile.ID).Msg("Profile created")

		// Set the profile as the current profile.
		profiles.SetCurrentProfile(pedal.instanceID, pedal.device, profile.ID)
	}

	currentProfile = profiles.GetCurrentProfile(pedal.instanceID, pedal.device)

	log.Info().Interface("current_profile", currentProfile).Msg("Current profile")

	// Buffer for outgoing events.
	buf := make([]byte, 512)

	// Get NATS connection an KV store.
	nc, kv := natsconn.GetNATSConn()

	// Listen for incoming device input.
	for {
		n, err := pedal.device.Read(buf)

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
					log.Debug().Interface("device", pedal.device).Int("button_index", buttonIndex).Msg("Button released")
					continue
				}

				log.Debug().Interface("device", pedal.device).Int("button_index", buttonIndex).Msg("Button pressed")

				key := "instances." + pedal.instanceID + ".devices." + pedal.device.Serial + ".profiles." + currentProfile.ID + ".switches." + strconv.Itoa(buttonIndex)

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
