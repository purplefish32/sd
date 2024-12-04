package pages

import (
	"encoding/json"
	"fmt"
	natsconn "sd/nats"
	"sd/streamdeck/xl/buttons"
	"sync"

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

func GetCurrentPage(instanceId string, device *hid.Device, profileId string) (*Page, error) {
	_, kv := natsconn.GetNATSConn()

	// Define the key for the current page
	key := "instances." + instanceId + ".devices." + device.Serial + ".profiles." + profileId + ".pages.current"

	// Get current page and page
	entry, err := kv.Get(key)

	if err != nil {
		if err == nats.ErrKeyNotFound {
			log.Printf("No current page found for device: %s", device.Serial)

			return nil, nil
		}
		log.Printf("Failed to get current page for device: %s, error: %v", device.Serial, err)

		return nil, err
	}

	// Parse the value into a Page struct
	var page Page

	if err := json.Unmarshal(entry.Value(), &page); err != nil {
		log.Error().Err(err).Msg("Failed to parse JSON")

		return nil, err
	}

	log.Info().
		Str("instance_id", instanceId).
		Str("device_serial", device.Serial).
		Str("profile_id", profileId).
		Str("page_id", page.ID).
		Msg("Current page found")

	return &page, nil
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

func CreatePage(instanceId string, device *hid.Device, profileId string) (pageId string, err error) {
	_, kv := natsconn.GetNATSConn()
	log.Printf("Creating Page for Instance: %v, device: %v, profile: %v", instanceId, device.Serial, profileId)

	// Generate a new UUID
	id := uuid.New()
	idStr := id.String()

	// Define the key for the current profile
	key := "instances." + instanceId + ".devices." + device.Serial + ".profiles." + profileId + ".pages." + idStr

	// Define a new page.
	page := Page{
		ID: idStr,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(page)

	if err != nil {
		log.Printf("Failed to serialize page data: %v", err)
		return "", err
	}

	// Put the serialized data into the KV store
	_, err = kv.Create(key, data)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Page key already exists: %s", key)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", key, err)
		}
		return "", err
	}

	// Create the default button.
	// button := buttons.Button{
	// 	Plugin: "",
	// 	Action: "",
	// 	Image: "./assets/images/black.jpg",
	// 	MetaData: buttons.ButtonMetadata{},
	// }

	// Convert button struct to JSON byte array
	//buttonData, err := json.Marshal(button)

	// if err != nil {
	// 	log.Fatal("Error Marshaling button:", err)
	// }

	// Create all missing keys in JetStream KV.
	var wg sync.WaitGroup // To wait for all goroutines to finish.

	for i := 0; i < 32; i++ {
		// Increment the WaitGroup counter.
		wg.Add(1)

		// Generate the key for the current iteration.
		//key := "instances." + instanceId + ".devices." + device.Serial + ".profiles." + profileId + ".pages." + pageId + ".buttons." + fmt.Sprintf("%v", i)

		// Launch a goroutine for each iteration.
		go func(i int) {
			defer wg.Done() // Decrement the counter when the goroutine finishes.

			buttons.CreateButton(instanceId, device, profileId, idStr, fmt.Sprintf("%v", i))

			// // Get the key-value entry.
			// entry, err := kv.Get(key)
			// if err != nil {
			// 	log.Printf("Error getting key %s: %v", key, err)
			// 	return
			// }

			// // Create the value (you can replace `buf` with the actual data you want to store).
			// _, err = kv.Create(key, buttonData)
			// if err != nil {
			// 	log.Printf("Error creating key %s: %v", key, err)
			// 	return
			// }

			// // Use the buffer (or call your utility function).
			// utils.SetKeyFromBuffer(device, i, entry.Value())
		}(i)
	}

	// Wait for all goroutines to complete.
	wg.Wait()

	log.Printf("Page created successfully: %+v", page)

	// Return the page UUID.
	return idStr, nil
}
