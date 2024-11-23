package pages

import (
	"encoding/json"
	"fmt"
	"log"
	"sd/streamdeck/xl/buttons"
	"sync"

	"github.com/google/uuid"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

type Page struct {
	ID string `json:"id"` // Unique identifier for the page
}

type CurrentPage struct {
	ID string `json:"id"` // Unique identifier for the profile
}

func CreatePage(kv nats.KeyValue, instanceId string, device *hid.Device, profileId string) (pageId string, err error) {
	log.Printf("Creating Page for Instance: %v, device: %v, profile: %v", instanceId, device.Serial, profileId)

	// Generate a new UUID
	id := uuid.New()
	idStr := id.String()

	log.Printf("IDSTR: %v", idStr)

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

			buttons.CreateButton(kv, instanceId, device, profileId, idStr, fmt.Sprintf("%v", i))

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

func SetCurrentPage(kv nats.KeyValue, instanceId string, device *hid.Device, profileId string, pageId string) error {
	// Define the key for the current page
	key := "instances." + instanceId + ".devices." + device.Serial + ".profiles." + profileId + "pages.current"

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