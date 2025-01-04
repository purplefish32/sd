package pages

import (
	"encoding/json"
	"sd/pkg/natsconn"

	"github.com/google/uuid"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Page struct {
	ID string `json:"id"` // Unique identifier for the page
}

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

func SetCurrentPage(instanceId string, device *hid.Device, profileId string, pageId string) error {
	_, kv := natsconn.GetNATSConn()

	log.Printf("Setting current page for profile: %v", profileId)

	// Define the key for the current page
	key := "instances." + instanceId + ".devices." + device.Serial + ".profiles." + profileId + ".pages.current"

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
		log.Printf("Failed to set current page for device: %s, error: %v", device.Serial, err) // TODO enrich the log with the rest of the data.
		return err
	}

	log.Printf("Current page set successfully for device: %s", device.Serial)

	return nil
}

func CreatePage(instanceId string, device *hid.Device, profileId string) (page Page) {
	_, kv := natsconn.GetNATSConn()
	log.Printf("Creating Page for Instance: %v, device: %v, profile: %v", instanceId, device.Serial, profileId)

	// Generate a new UUID
	id := uuid.New()
	idStr := id.String()

	// Define the key for the current profile
	key := "instances." + instanceId + ".devices." + device.Serial + ".profiles." + profileId + ".pages." + idStr

	// Define a new page.
	p := Page{
		ID: idStr,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(p)

	if err != nil {
		log.Printf("Failed to serialize page data: %v", err)
		return Page{}
	}

	// Put the serialized data into the KV store
	_, err = kv.Create(key, data)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Page key already exists: %s", key)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", key, err)
		}
		return Page{}
	}

	log.Printf("Page created successfully: %+v", page)

	// Return the page.
	return p
}
