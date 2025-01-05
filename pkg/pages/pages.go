package pages

import (
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"sd/pkg/store"
	"strings"

	"sd/pkg/types"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Page = types.Page

type CurrentPage struct {
	ID string `json:"id"` // Unique identifier for the profile
}

func GetCurrentPage(instanceId string, deviceId string, profileId string) *Page {
	_, kv := natsconn.GetNATSConn()

	// Define the key for the current page
	key := "instances." + instanceId + ".devices." + deviceId + ".profiles." + profileId + ".pages.current"

	// Get current page and page
	entry, err := kv.Get(key)

	if err != nil {
		if err == nats.ErrKeyNotFound {
			log.Printf("No current page found for device: %s", deviceId)

			return nil
		}
		log.Printf("Failed to get current page for device: %s, error: %v", deviceId, err)

		return nil
	}

	// Parse the value into a Page struct
	var page Page

	if err := json.Unmarshal(entry.Value(), &page); err != nil {
		log.Error().Err(err).Msg("Failed to parse JSON")

		return nil
	}

	log.Info().
		Str("instance_id", instanceId).
		Str("device_serial", deviceId).
		Str("profile_id", profileId).
		Str("page_id", page.ID).
		Msg("Current page found")

	return &page
}

func SetCurrentPage(instanceId string, deviceId string, profileId string, pageId string) error {
	_, kv := natsconn.GetNATSConn()

	log.Printf("Setting current page for profile: %v", profileId)

	// Define the key for the current page
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.current",
		instanceId, deviceId, profileId)

	log.Printf("KEY: %v", key)
	currentPage := CurrentPage{
		ID: pageId,
	}

	// Serialize the Page struct to JSON
	data, err := json.Marshal(currentPage)
	if err != nil {
		log.Printf("Failed to serialize page data: %v", err)
		return err
	}

	// Put the serialized data into the KV store
	if _, err := kv.Put(key, data); err != nil {
		log.Printf("Failed to set current page for device: %s, error: %v", deviceId, err) // TODO enrich the log with the rest of the data.
		return err
	}

	log.Printf("Current page set successfully for device: %s", deviceId)

	return nil
}

func CreatePage(instanceID string, deviceID string, profileID string) (Page, error) {
	_, kv := natsconn.GetNATSConn()
	log.Printf("Creating Page for Instance: %v, device: %v, profile: %v", instanceID, deviceID, profileID)

	// Generate a new UUID
	id := uuid.New()
	idStr := id.String()

	// Define the key for the current profile
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s",
		instanceID, deviceID, profileID, idStr)

	// Define a new page
	p := Page{
		ID: idStr,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(p)
	if err != nil {
		return Page{}, fmt.Errorf("failed to serialize page data: %w", err)
	}

	// Put the serialized data into the KV store
	_, err = kv.Put(key, data)
	if err != nil {
		return Page{}, fmt.Errorf("failed to create page in KV store: %w", err)
	}

	log.Printf("Page created successfully: %+v", p)

	// After successfully creating the page, update the profile
	profile, err := store.GetProfile(instanceID, deviceID, profileID)
	if err != nil {
		return p, fmt.Errorf("failed to get profile: %w", err)
	}

	// Add the new page to the profile's pages array
	profile.Pages = append(profile.Pages, p)

	// Save the updated profile
	err = store.UpdateProfile(instanceID, deviceID, profileID, profile)
	if err != nil {
		return p, fmt.Errorf("failed to update profile with new page: %w", err)
	}

	return p, nil
}

func GetPages(instanceId string, deviceId string, profileId string) ([]Page, error) {
	_, kv := natsconn.GetNATSConn()

	// Define the key prefix to search for pages
	prefix := "instances." + instanceId + ".devices." + deviceId + ".profiles." + profileId + ".pages."

	// List the keys in the NATS KV store under the given prefix
	keyLister, err := kv.ListKeys()
	if err != nil {
		log.Error().Err(err).Msg("Could not list NATS KV keys")
		return nil, err
	}

	// Initialize a slice to store the pages
	var pages []Page

	// Iterate over the keys from the channel
	for key := range keyLister.Keys() {
		// If the key doesn't start with the prefix, skip it
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		// Skip the current page key
		if strings.HasSuffix(key, ".current") {
			continue
		}

		// Ensure we're not getting nested items
		if strings.Count(key[len(prefix):], ".") > 0 {
			continue
		}

		// Fetch the page data for each key
		val, err := kv.Get(key)
		if err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to get value for key")
			continue
		}

		// Parse the page data
		var page Page
		err = json.Unmarshal(val.Value(), &page)
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
