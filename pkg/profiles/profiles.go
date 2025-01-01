package profiles

import (
	"encoding/json"
	"sd/pkg/natsconn"
	"sd/pkg/pages"

	"github.com/google/uuid"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Profile struct {
	ID    string       `json:"id"`    // Unique identifier for the profile
	Name  string       `json:"name"`  // Display name for the profile
	Pages []pages.Page `json:"pages"` // List of pages in the profile
	//Default int     `json:"default"` // Index of the default page
}

type CurrentProfile struct {
	ID string `json:"id"` // Unique identifier for the profile
}

// type Button struct {
// 	ID          int      `json:"id"`          // Button ID (matches physical button on the stream deck)
// 	ActionType  string   `json:"action_type"` // Type of action (e.g., "update_image", "change_profile")
// 	ActionValue string   `json:"action_value"`// Value associated with the action (e.g., image ID, profile ID)
// 	Labels      []string `json:"labels"`      // Optional labels for multi-state buttons
// 	Image       string   `json:"image"`       // Key to the image in NATS KV
// }

// TODO UpdatePage, DeletePage

func CreateProfile(instanceID string, device *hid.Device, name string) (profile Profile, err error) {
	_, kv := natsconn.GetNATSConn()

	log.Printf("Creating Profile for Instance: %v, device: %v", instanceID, device.Serial)

	p := Profile{
		ID:   uuid.New().String(),
		Name: name,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(p)

	if err != nil {
		log.Printf("Failed to serialize profile data: %v", err)
		return Profile{}, err
	}

	// Define the key for the current profile
	key := "instances." + instanceID + ".devices." + device.Serial + ".profiles." + p.ID

	// Put the serialized data into the KV store
	_, err = kv.Create(key, data)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Profile key already exists: %s", key)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", key, err)
		}
		return Profile{}, err
	}

	// Set page as default page.
	return p, nil
}

func GetCurrentProfile(instanceId string, device *hid.Device) *Profile { // TODO move this to the device.
	_, kv := natsconn.GetNATSConn()

	// Define the key for the current profile
	key := "instances." + instanceId + ".devices." + device.Serial + ".profiles.current"

	// Get current profile and page
	entry, err := kv.Get(key)

	if err != nil {
		if err == nats.ErrKeyNotFound {
			log.Error().Err(err).Str("device_serial", device.Serial).Msg("No NATS key for current profile found")
			return nil
		}
		log.Error().Err(err).Str("device_serial", device.Serial).Msg("Failed to get current profile")
		return nil
	}

	// Parse the value into a Profile struct
	var profile Profile

	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		log.Error().Err(err).Msg("Failed to parse JSON")
		return nil
	}

	log.Info().
		Str("instance_id", instanceId).
		Str("device_serial", device.Serial).
		Str("profile_id", profile.ID).
		Msg("Current profile found")

	return &profile
}

func SetCurrentProfile(instanceId string, deviceId string, profileId string) error {
	_, kv := natsconn.GetNATSConn()

	// Define the key for the current profile
	key := "instances." + instanceId + ".devices." + deviceId + ".profiles.current"

	currentProfile := CurrentProfile{
		ID: profileId,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(currentProfile)

	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize json")
		return err
	}

	// Put the serialized data into the KV store
	if _, err := kv.Put(key, data); err != nil {
		log.Error().
			Str("instance_id", instanceId).
			Str("device_id", deviceId).
			Err(err).
			Msg("Failed to set current profile")

		return err
	}

	log.Info().
		Str("instance_id", instanceId).
		Str("device_id", deviceId).
		Msg("Current profile set successfully")

	return nil
}
