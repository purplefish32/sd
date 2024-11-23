package profiles

import (
	"encoding/json"
	"log"
	"sd/streamdeck/xl/pages"

	"github.com/google/uuid"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

type Profile struct {
	ID      string       `json:"id"`      // Unique identifier for the profile
	Name    string       `json:"name"`    // Display name for the profile
	Pages   []pages.Page `json:"pages"`   // List of pages in the profile
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


func CreateProfile(kv nats.KeyValue, instanceID string, device *hid.Device, name string) (profileId string, err error) {
	log.Printf("Creating Profile for Instance: %v, device: %v", instanceID, device.Serial)
	
	// Generate a new UUID
	id := uuid.New()
	idStr := id.String()

	// Define the key for the current profile
	key := "instances." + instanceID + ".devices." + device.Serial + ".profiles." + idStr

	profile := Profile{
		ID: idStr,
		Name: name,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(profile)

	if err != nil {
		log.Printf("Failed to serialize profile data: %v", err)
		return "", err
	}

	// Put the serialized data into the KV store
	_, err = kv.Create(key, data)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Profile key already exists: %s", key)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", key, err)
		}
		return "", err
	}

	// Create new page.
	pages.CreatePage(kv, instanceID, device, profile.ID)

	log.Printf("Profile created successfully: %+v", profile)

	// Set page as default page.
	return idStr, nil
}

func GetCurrentProfile(kv nats.KeyValue, instanceID string, device *hid.Device) (*Profile, error) {
	// Define the key for the current profile
	key := "instances." + instanceID + ".devices." + device.Serial + ".profiles.current"

	// Get current profile and page
	entry, err := kv.Get(key)

	if err != nil {
		if err == nats.ErrKeyNotFound {
			log.Printf("No current profile found for device: %s", device.Serial)
			return nil, nil
		}
		log.Printf("Failed to get current profile for device: %s, error: %v", device.Serial, err)
		return nil, err
	}

	// Parse the value into a Profile struct
	var profile Profile

	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		log.Printf("Failed to parse profile data: %v", err)
		return nil, err
	}

	log.Printf("Current profile retrieved: %+v", profile)
	return &profile, nil
}

func SetCurrentProfile(kv nats.KeyValue, instanceID string, device *hid.Device, profileId string) error {
	// Define the key for the current profile
	key := "instances." + instanceID + ".devices." + device.Serial + ".profiles.current"

	currentProfile := CurrentProfile{
		ID: profileId,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(currentProfile)
	if err != nil {
		log.Printf("Failed to serialize profile data: %v", err)
		return err
	}

	// Put the serialized data into the KV store
	if _, err := kv.Put(key, data); err != nil {
		log.Printf("Failed to set current profile for device: %s, error: %v", device.Serial, err)
		return err
	}

	log.Printf("Current profile set successfully for device: %s", device.Serial)

	return nil
}