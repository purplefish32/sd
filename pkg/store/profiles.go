package store

import (
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"sd/pkg/types"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

func GetProfiles(instanceID string, deviceID string) []types.Profile {
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
		profiles = append(profiles, profile)
	}

	return profiles
}

func CreateProfile(instanceID, deviceID, name string) (*types.Profile, error) {
	if instanceID == "" || deviceID == "" || name == "" {
		return nil, fmt.Errorf("instanceID, deviceID, and name are required")
	}

	profile := &types.Profile{
		ID:          uuid.New().String(),
		Name:        name,
		Pages:       make([]types.Page, 0), // Initialize empty slice
		CurrentPage: "",                    // Explicit empty string
	}

	// Save the profile
	if err := UpdateProfile(instanceID, deviceID, profile.ID, profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

func GetCurrentProfile(instanceID string, deviceID string) types.Profile {
	log.Info().Str("instanceID", instanceID).Str("deviceID", deviceID).Msg("Getting current profile")
	_, kv := natsconn.GetNATSConn()

	// Define the key for the current profile
	key := "instances." + instanceID + ".devices." + deviceID

	// Get the device
	entry, err := kv.Get(key)
	if err != nil {
		log.Warn().Str("device_serial", deviceID).Msg("No NATS key for current profile found")
		return types.Profile{}
	}

	// Parse the value into a Profile struct
	var device types.Device
	if err := json.Unmarshal(entry.Value(), &device); err != nil {
		log.Warn().Err(err).Msg("Failed to unmarshal device")
		return types.Profile{}
	}

	if device.CurrentProfile == "" {
		log.Warn().Msg("No current profile found")
		return types.Profile{}
	}

	profile := GetProfile(instanceID, deviceID, device.CurrentProfile)
	return profile
}

func SetCurrentProfile(instanceID string, deviceID string, profileID string) error {
	log.Info().Str("instanceID", instanceID).Str("deviceID", deviceID).Str("profileID", profileID).Msg("Setting current profile")
	_, kv := natsconn.GetNATSConn()

	// Define the key for the current device
	key := "instances." + instanceID + ".devices." + deviceID

	log.Info().Str("key", key).Msg("Key")

	// Get the device
	entry, err := kv.Get(key)

	log.Info().Interface("entry", entry).Msg("Entry")

	if err != nil {
		if err == nats.ErrKeyNotFound {
			log.Warn().Str("device_serial", deviceID).Msg("No NATS key for current profile found")
		}
		return nil
	}

	var device types.Device

	if err := json.Unmarshal(entry.Value(), &device); err != nil {
		return nil
	}

	device.CurrentProfile = profileID

	log.Info().Interface("device", device).Msg("Device")

	// Save the device to NATS
	deviceData, err := json.Marshal(device)
	if err != nil {
		return err
	}

	log.Info().Interface("deviceData", deviceData).Msg("Device data")

	kv.Put(key, deviceData)

	log.Info().Str("key", key).Msg("Key")

	return nil
}

func GetProfile(instanceID, deviceID, profileID string) types.Profile {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profileID)

	entry, err := kv.Get(key)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get profile")
		return types.Profile{}
	}

	var profile types.Profile
	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		log.Warn().Err(err).Msg("Failed to unmarshal profile")
		return types.Profile{}
	}

	return profile
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
	log.Info().Str("instanceId", instanceID).Str("deviceId", deviceID).Str("profileId", profileID).Msg("Deleting profile")
	_, kv := natsconn.GetNATSConn()

	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profileID)

	// All pages in the profile need to be deleted too.
	p := GetPages(instanceID, deviceID, profileID)

	log.Info().Interface("pages", p).Msg("Pages")

	for _, page := range p {
		log.Info().Str("page_id", page.ID).Msg("Deleting page")
		DeletePage(instanceID, deviceID, profileID, page.ID)
	}

	log.Info().Str("key", key).Msg("Deleting profile")
	kv.Delete(key)

	return nil
}
