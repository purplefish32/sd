package profiles

import (
	"encoding/json"
	"fmt"
	"sd/pkg/env"
	"sd/pkg/natsconn"
	"sd/pkg/store"
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
}

type CurrentProfile struct {
	ID string `json:"id"` // Unique identifier for the profile
}

func GetProfiles(instanceId string, deviceId string) ([]Profile, error) {
	_, kv := natsconn.GetNATSConn()

	// Define the key prefix to search for profiles
	prefix := "instances." + instanceId + ".devices." + deviceId + ".profiles."

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

func CreateProfile(instanceID, deviceID, name string) (*types.Profile, error) {
	profile := &types.Profile{
		ID:   uuid.New().String(),
		Name: name,
		TouchScreen: types.TouchScreenLayout{
			Mode:      "full",
			FullImage: env.Get("ASSET_PATH", "") + "images/black.png",
			Segments:  [4]string{},
		},
	}

	// Save the profile
	if err := store.UpdateProfile(instanceID, deviceID, profile.ID, profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

func GetCurrentProfile(instanceID string, deviceID string) *types.Profile {
	_, kv := natsconn.GetNATSConn()

	// Define the key for the current profile
	key := "instances." + instanceID + ".devices." + deviceID + ".profiles.current"

	// Get current profile and page
	entry, err := kv.Get(key)

	if err != nil {
		if err == nats.ErrKeyNotFound {
			log.Warn().Str("device_serial", deviceID).Msg("No NATS key for current profile found")
		}
		return nil
	}

	// Parse the value into a Profile struct
	var profile types.Profile

	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		return nil
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("device_serial", deviceID).
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
