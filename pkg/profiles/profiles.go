package profiles

import (
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"sd/pkg/pages"
	"sd/pkg/store"
	"sd/pkg/types"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type CurrentProfile struct {
	ID string `json:"id"` // Unique identifier for the profile
}

func GetProfiles(instanceID string, deviceID string) ([]types.Profile, error) {
	_, kv := natsconn.GetNATSConn()

	// List the keys in the NATS KV store under the given prefix
	keyList, err := kv.ListKeys()

	if err != nil {
		log.Error().Err(err).Msg("Could not list NATS KV keys")
		return nil, err
	}

	// Initialize a slice to store the profiles
	var profiles []types.Profile

	// Iterate over the keys from the channel
	for key := range keyList.Keys() {

		// If the key doesn't start with the prefix, skip it
		if !strings.Contains(key, fmt.Sprintf("instances.%s.devices.%s.profiles.", instanceID, deviceID)) {
			continue
		}

		parts := strings.Split(key, ".")

		if len(parts) != 6 {
			continue
		}

		// Skip the current profile
		// TODO: This is a hack to skip the current profile, we should find a better way to do this
		if parts[5] == "current" {
			log.Info().Str("key", key).Msg("Skipping current profile")
			continue
		}

		// Fetch the profile data for each key
		entry, err := kv.Get(key)

		if err != nil {
			log.Warn().Err(err).Str("key", key).Msg("Skipping invalid device entry")
			continue
		}

		// Assuming the profile data is stored as a JSON string or similar structure
		var profile types.Profile

		err = json.Unmarshal(entry.Value(), &profile)
		if err != nil {
			log.Warn().Err(err).Str("key", key).Msg("Skipping malformed device data")
			continue
		}

		// Append the profile to the list
		profiles = append(profiles, types.Profile{
			ID:    profile.ID,
			Name:  profile.Name,
			Pages: profile.Pages,
		})
	}

	return profiles, nil
}

func CreateProfile(instanceID, deviceID, name string) (*types.Profile, error) {
	profile := &types.Profile{
		ID:   uuid.New().String(),
		Name: name,
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

func GetProfile(instanceID, deviceID, profileID string) (*types.Profile, error) {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profileID)

	entry, err := kv.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	var profile types.Profile
	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

func UpdateProfile(instanceID, deviceID, profileID string, profile *types.Profile) error {
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

func DeleteProfile(instanceID, deviceID, profileID string) error {
	_, kv := natsconn.GetNATSConn()

	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profileID)

	kv.Delete(key)

	// All pages in the profile need to be deleted too.
	pages, err := pages.GetPages(instanceID, deviceID, profileID)

	if err != nil {
		return err
	}

	for _, page := range pages {
		//pages.DeletePage(instanceID, deviceID, profileID, page.ID)
		log.Info().Str("page_id", page.ID).Msg("Deleting page")
	}

	return nil
}
