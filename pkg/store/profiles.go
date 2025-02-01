package store

import (
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"sd/pkg/types"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func GetProfiles(instanceID string, device *types.Device) []types.Profile {
	_, kv := natsconn.GetNATSConn()

	// List the keys in the NATS KV store under the given prefix
	keyList, err := kv.ListKeys()

	if err != nil {
		log.Error().Err(err).Msg("Could not list NATS KV keys")
		return nil
	}

	// Initialize a slice to store the profiles
	var profiles []types.Profile

	// Iterate over the keys from the channel
	for key := range keyList.Keys() {

		// If the key doesn't start with the prefix, skip it
		if !strings.Contains(key, fmt.Sprintf("instances.%s.devices.%s.profiles.", instanceID, device.ID)) {
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
		profiles = append(profiles, profile)
	}

	return profiles
}

func CreateProfile(instanceID string, device *types.Device, name string) (*types.Profile, error) {
	// Input Validation: Ensuring no parameter is an empty string
	if instanceID == "" || device == nil || name == "" {
		return nil, fmt.Errorf("instanceID, device, and name are required")
	}

	var blankProfile = types.Profile{
		ID:          uuid.New().String(),
		Name:        name,
		Pages:       make([]types.Page, 0),
		CurrentPage: "", // Explicit empty string
	}

	// Save the profile
	profile, err := UpdateProfile(instanceID, device, &blankProfile)

	if err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	// Create a blank page for the profile
	page, err := CreatePage(instanceID, device, profile.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	profile.Pages = append(profile.Pages, *page)
	profile.CurrentPage = page.ID

	return profile, nil
}

// func SetCurrentProfile(instanceID string, deviceID string, profileID string) error {
// 	log.Info().Str("instanceID", instanceID).Str("deviceID", deviceID).Str("profileID", profileID).Msg("Setting current profile")
// 	_, kv := natsconn.GetNATSConn()

// 	// Define the key for the current device
// 	key := "instances." + instanceID + ".devices." + deviceID

// 	log.Info().Str("key", key).Msg("Key")

// 	// Get the device
// 	entry, err := kv.Get(key)

// 	log.Info().Interface("entry", entry).Msg("Entry")

// 	if err != nil {
// 		if err == nats.ErrKeyNotFound {
// 			log.Warn().Str("device_serial", deviceID).Msg("No NATS key for current profile found")
// 		}
// 		return nil
// 	}

// 	var device types.Device

// 	if err := json.Unmarshal(entry.Value(), &device); err != nil {
// 		return nil
// 	}

// 	device.CurrentProfile = profileID

// 	log.Info().Interface("device", device).Msg("Device")

// 	// Save the device to NATS
// 	deviceData, err := json.Marshal(device)
// 	if err != nil {
// 		return err
// 	}

// 	log.Info().Interface("deviceData", deviceData).Msg("Device data")

// 	kv.Put(key, deviceData)

// 	log.Info().Str("key", key).Msg("Key")

// 	return nil
// }

func GetProfile(instanceID string, device *types.Device, profileID string) *types.Profile {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, device.ID, profileID)

	entry, err := kv.Get(key)
	if err != nil {
		log.Warn().Err(err).Str("key", key).Msg("Failed to get profile")
		return nil
	}

	var profile types.Profile
	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		log.Warn().Err(err).Msg("Failed to unmarshal profile")
		return nil
	}

	return &profile
}

func UpdateProfile(instanceID string, device *types.Device, profile *types.Profile) (*types.Profile, error) {
	// Input validation for instanceID, deviceID, profileID, and profile pointer
	if instanceID == "" || device == nil || profile == nil {
		return nil, fmt.Errorf("instanceID, deviceID, and profile are required")
	}

	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, device.ID, profile.ID)

	json, err := json.Marshal(profile)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile: %w", err)
	}

	_, err = kv.Put(key, json)
	if err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

func DeleteProfile(instanceID string, device *types.Device, profileID string) error {
	// Input validation for instanceID, deviceID, profileID, and profile pointer
	if instanceID == "" || device == nil || profileID == "" {
		return fmt.Errorf("instanceID, deviceID and profileID are required")
	}

	log.Info().Str("instanceId", instanceID).Interface("device", device).Str("profileId", profileID).Msg("Deleting profile")

	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, device.ID, profileID)

	// All pages in the profile need to be deleted too.
	p := GetPages(instanceID, device, profileID)
	log.Info().Interface("pages", p).Msg("Pages")

	for _, page := range p {
		log.Info().Str("page_id", page.ID).Msg("Deleting page")
		if err := DeletePage(instanceID, device, profileID, page.ID); err != nil {
			log.Warn().Err(err).Str("page_id", page.ID).Msg("Error deleting page, continuing")
		}
	}

	log.Info().Str("key", key).Msg("Deleting profile")

	err := kv.Delete(key)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}
