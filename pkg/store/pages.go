package store

import (
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"strings"

	"sd/pkg/types"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

func GetCurrentPage(instanceId string, deviceId string, profileId string) *types.Page {
	_, kv := natsconn.GetNATSConn()

	// Define the key for the current profile
	key := "instances." + instanceId + ".devices." + deviceId + ".profiles." + profileId

	// Get current profile
	entry, err := kv.Get(key)

	if err != nil {
		if err == nats.ErrKeyNotFound {
			log.Printf("Device key not found: %s", deviceId)
			return nil
		}
		log.Printf("Failed to get device: %s, error: %v", deviceId, err)
		return nil
	}

	// Parse the value into a Page struct
	var profile types.Profile

	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		log.Error().Err(err).Msg("Failed to parse JSON")
		return nil
	}

	key = "instances." + instanceId + ".devices." + deviceId + ".profiles." + profileId + ".pages." + profile.CurrentPage

	entry, err = kv.Get(key)

	if err != nil {
		if err == nats.ErrKeyNotFound {
			log.Printf("Page key not found: %s", deviceId)

			return nil
		}
		log.Printf("Failed to get page: %s, error: %v", deviceId, err)

		return nil
	}

	var page types.Page

	if err := json.Unmarshal(entry.Value(), &page); err != nil {
		log.Error().Err(err).Msg("Failed to parse JSON")

		return nil
	}

	return &page
}

func SetCurrentPage(instanceId string, deviceId string, profileId string, pageId string) error {
	_, kv := natsconn.GetNATSConn()

	// Define the key for the profile
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s",
		instanceId, deviceId, profileId)

	// Get the profile
	entry, err := kv.Get(key)

	if err != nil {
		return err
	}

	var profile types.Profile

	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		return err
	}

	profile.CurrentPage = pageId

	profileData, err := json.Marshal(profile)

	if err != nil {
		return err
	}

	kv.Put(key, profileData)

	return nil
}

func CreatePage(instanceID string, deviceID string, profileID string) (*types.Page, error) {
	_, kv := natsconn.GetNATSConn()
	log.Printf("Creating Page for Instance: %v, device: %v, profile: %v", instanceID, deviceID, profileID)

	// Generate a new UUID
	id := uuid.New()
	idStr := id.String()

	// Define the key for the current profile
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s",
		instanceID, deviceID, profileID, idStr)

	// Define a new page
	p := types.Page{
		ID: idStr,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize page data: %w", err)
	}

	// Put the serialized data into the KV store
	_, err = kv.Put(key, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create page in KV store: %w", err)
	}

	log.Printf("Page created successfully: %+v", p)

	// After successfully creating the page, update the profile
	profile := GetProfile(instanceID, deviceID, profileID)

	// Add the new page to the profile's pages array
	profile.Pages = append(profile.Pages, p)

	// Save the updated profile
	err = UpdateProfile(instanceID, deviceID, profileID, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile with new page: %w", err)
	}

	log.Printf("Page created successfully: %+v", p)

	return &p, nil
}

func GetPages(instanceId string, deviceId string, profileId string) ([]types.Page, error) {
	_, kv := natsconn.GetNATSConn()

	// Define the key prefix to search for pages
	var prefix = "instances." + instanceId + ".devices." + deviceId + ".profiles." + profileId + ".pages."

	// List the keys in the NATS KV store under the given prefix
	keyLister, err := kv.ListKeys()
	if err != nil {
		log.Error().Err(err).Msg("Could not list NATS KV keys")
		return nil, err
	}

	// Initialize a slice to store the pages
	var pages []types.Page

	// Iterate over the keys from the channel
	for key := range keyLister.Keys() {
		// If the key doesn't start with the prefix, skip it
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		// Ensure we're not getting nested items
		if strings.Count(key[len(prefix):], ".") > 0 {
			continue
		}

		// Fetch the page data for each key
		entry, err := kv.Get(key)
		if err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to get value for key")
			continue
		}

		// Parse the page data
		var page types.Page
		err = json.Unmarshal(entry.Value(), &page)
		if err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to unmarshal page data")
			continue
		}

		// Append the page to the list
		pages = append(pages, page)
	}

	log.Info().Interface("pages", pages).Msg("Retrieved pages")

	return pages, nil
}

func DeletePage(instanceID string, deviceID string, profileID string, pageID string) error {
	log.Info().Str("instanceId", instanceID).Str("deviceId", deviceID).Str("profileId", profileID).Str("pageId", pageID).Msg("Deleting page")
	_, kv := natsconn.GetNATSConn()

	// I want to delete all buttons in the page
	b, err := GetButtons(instanceID, deviceID, profileID, pageID)
	log.Info().Interface("buttons", b).Msg("Buttons")

	if err != nil {
		log.Error().Err(err).Msg("Failed to get buttons")
		return err
	}

	for _, button := range b {
		log.Info().Str("button_id", button.ID).Msg("Deleting button")
		DeleteButton(instanceID, deviceID, profileID, pageID, button.ID)
	}

	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s", instanceID, deviceID, profileID, pageID)

	log.Info().Str("key", key).Msg("Deleting page")

	kv.Delete(key)

	return nil
}
