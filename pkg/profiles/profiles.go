package profiles

import (
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"sd/pkg/pages"
	"sd/pkg/types"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Profile struct {
	ID    string       `json:"id"`    // Unique identifier for the profile
	Name  string       `json:"name"`  // Display name for the profile
	Pages []types.Page `json:"pages"` // List of pages in the profile
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

func GetProfiles(instanceId string, deviceId string) ([]Profile, error) {
	_, kv := natsconn.GetNATSConn()

	// Define the key prefix to search for profiles
	prefix := "instances." + instanceId + ".devices." + deviceId + ".profiles."

	// Define the key pattern to search for profiles
	//keyPrefix := "instances." + instanceId + ".devices." + deviceId + ".profiles."

	// List the keys in the NATS KV store under the given prefix
	keyLister, err := kv.ListKeys()
	if err != nil {
		log.Error().Err(err).Msg("Could not list NATS KV keys")
		return nil, err
	}

	// Initialize a slice to store the profiles
	var profiles []Profile

	// Iterate over the keys from the channel
	for key := range keyLister.Keys() {

		// If the key doesn't start with the prefix, skip it
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		// Ensure that the key ends with the profile UUID and doesn't contain additional parts like ".pages" or ".buttons"
		if strings.Contains(key[len(prefix):], ".") {
			continue // Ignore keys that contain further parts (like .pages or .buttons)
		}

		// If the key ends with ".current", skip it (it's just the default profile)
		if strings.HasSuffix(key, ".current") {
			continue
		}

		// Fetch the profile data for each key
		val, err := kv.Get(key)
		if err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to get value for key")
			continue
		}

		// Assuming the profile data is stored as a JSON string or similar structure
		var profile Profile
		err = json.Unmarshal(val.Value(), &profile)
		if err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to unmarshal profile data")
			continue
		}

		// Append the profile to the list
		profiles = append(profiles, profile)
	}

	log.Info().Interface("profiles", profiles).Msg("Key")

	return profiles, nil
}

func CreateProfile(instanceID string, deviceID string, name string) (*Profile, error) {
	profile := Profile{
		ID:   uuid.New().String(),
		Name: name,
	}

	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profile.ID)

	data, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile: %w", err)
	}

	_, err = kv.Put(key, data)
	if err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	// Create default page
	_, err = pages.CreatePage(instanceID, deviceID, profile.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create default page")
		return nil, err
	}

	return &profile, nil
}

func GetCurrentProfile(instanceId string, deviceId string) *Profile { // TODO move this to the device.
	_, kv := natsconn.GetNATSConn()

	// Define the key for the current profile
	key := "instances." + instanceId + ".devices." + deviceId + ".profiles.current"

	// Get current profile and page
	entry, err := kv.Get(key)

	if err != nil {
		if err == nats.ErrKeyNotFound {
			log.Error().Err(err).Str("device_serial", deviceId).Msg("No NATS key for current profile found")
			return nil
		}
		log.Error().Err(err).Str("device_serial", deviceId).Msg("Failed to get current profile")
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
		Str("device_serial", deviceId).
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

func GetProfile(instanceID, deviceID, profileID string) (*Profile, error) {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profileID)

	entry, err := kv.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

func UpdateProfile(instanceID, deviceID, profileID string, profile *Profile) error {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profileID)

	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	_, err = kv.Put(key, data)
	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}
