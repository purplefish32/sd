package store

import (
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"strconv"
	"strings"

	"sd/pkg/types"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func GetPage(instanceID, deviceID, profileID, pageID string) *types.Page {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s", instanceID, deviceID, profileID, pageID)

	entry, err := kv.Get(key)
	if err != nil {
		return nil
	}

	var page types.Page
	if err := json.Unmarshal(entry.Value(), &page); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal page")
		return nil
	}

	return &page
}

func CreatePage(instanceID string, device *types.Device, profileID string) (*types.Page, error) {
	_, kv := natsconn.GetNATSConn()
	log.Printf("Creating Page for Instance: %v, device: %v, profile: %v", instanceID, device.ID, profileID)

	// Generate a new UUID
	id := uuid.New()
	idStr := id.String()

	// Define the key for the current profile
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s",
		instanceID, device.ID, profileID, idStr)

	// Define a new page
	newPage := types.Page{
		ID: idStr,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(newPage)

	if err != nil {
		return nil, fmt.Errorf("failed to serialize page data: %w", err)
	}

	// Put the serialized data into the KV store
	_, err = kv.Put(key, data)

	if err != nil {
		return nil, fmt.Errorf("failed to create page in KV store: %w", err)
	}

	// After creating the page, update the profile
	profile := GetProfile(instanceID, device, profileID)
	profile.Pages = append(profile.Pages, types.Page{ID: newPage.ID})
	profile.CurrentPage = newPage.ID

	// Update profile in KV store
	profileData, err := json.Marshal(profile)

	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal profile")
		return &newPage, err
	}

	_, err = kv.Put(fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, device.ID, profileID), profileData)

	if err != nil {
		log.Error().Err(err).Msg("Failed to update profile")
		return &newPage, err
	}

	log.Info().Str("device_type", device.Type).Msg("device.Type")

	if device.Type == "xl" {
		// Create 32 buttons for the page
		for i := 0; i < 32; i++ {
			CreateButton(instanceID, device, profileID, newPage.ID, strconv.Itoa(i+1))
		}
	}

	return &newPage, nil
}

func GetPages(instanceID string, device *types.Device, profileID string) []types.Page {
	_, kv := natsconn.GetNATSConn()

	// Define the key prefix to search for pages
	var prefix = "instances." + instanceID + ".devices." + device.ID + ".profiles." + profileID + ".pages."

	// List the keys in the NATS KV store under the given prefix
	keyLister, err := kv.ListKeys()
	if err != nil {
		log.Error().Err(err).Msg("Could not list NATS KV keys")
		return nil
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

	return pages
}

func DeletePage(instanceID string, device *types.Device, profileID string, pageID string) error {
	_, kv := natsconn.GetNATSConn()

	// Get all buttons in the page
	b, err := GetButtons(instanceID, device, profileID, pageID)
	log.Info().Interface("buttons", b).Msg("Buttons")

	if err != nil {
		log.Error().Err(err).Msg("Failed to get buttons")
		return err
	}

	// Delete all buttons in the page
	for _, button := range b {
		log.Info().Str("button_id", button.ID).Msg("Deleting button")
		DeleteButton(instanceID, device, profileID, pageID, button.ID)
	}

	// Delete the page
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s", instanceID, device.ID, profileID, pageID)
	log.Info().Str("key", key).Msg("Deleting page")
	kv.Delete(key)

	// Update the profile
	profile := GetProfile(instanceID, device, profileID)

	// Remove the page from the profile

	if len(profile.Pages) > 0 {
		for i, p := range profile.Pages {
			if p.ID == pageID {
				profile.Pages = append(profile.Pages[:i], profile.Pages[i+1:]...)

				// If the page is removed the current page should be the previous page
				if i > 1 {
					profile.CurrentPage = profile.Pages[i-1].ID
				} else {
					profile.CurrentPage = profile.Pages[0].ID // TODO there is a bug here
				}

				break
			}
		}
	} else {
		page, err := CreatePage(instanceID, device, profileID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create page")
			return err
		}
		profile.CurrentPage = page.ID
	}

	UpdateProfile(instanceID, device, profile)

	log.Info().Interface("profile", profile).Msg("Profile")

	return nil
}
