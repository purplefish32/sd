package store

import (
	"encoding/json"
	"fmt"
	"sd/pkg/env"
	"sd/pkg/natsconn"
	"sd/pkg/types"
	"strings"

	"github.com/h2non/bimg"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// type Title struct {
// }

func GetButton(key string) (button types.Button, err error) {
	_, kv := natsconn.GetNATSConn()

	entry, err := kv.Get(key)
	if err != nil {
		log.Error().Str("key", key).Err(err).Msg("Failed to retrieve key")
		return types.Button{}, err
	}

	err = json.Unmarshal(entry.Value(), &button)

	if err != nil {
		log.Error().Err(err).Msg("Failed to parse JSON")
		return types.Button{}, err
	}

	return button, nil
}

func GetButtons(instanceID string, deviceID string, profileID string, pageID string) ([]types.Button, error) {
	log.Info().Str("instanceId", instanceID).Str("deviceId", deviceID).Str("profileId", profileID).Str("pageId", pageID).Msg("Getting buttons")
	_, kv := natsconn.GetNATSConn()

	// Define the key prefix to search for buttons
	var prefix = "instances." + instanceID + ".devices." + deviceID + ".profiles." + profileID + ".pages." + pageID + ".buttons."
	log.Info().Str("prefix", prefix).Msg("Prefix")

	// List the keys in the NATS KV store under the given prefix
	keyLister, err := kv.ListKeys()
	log.Info().Interface("keyLister", keyLister).Msg("KeyLister")

	if err != nil {
		log.Error().Err(err).Msg("Could not list NATS KV keys")
		return nil, err
	}

	var buttons []types.Button

	for key := range keyLister.Keys() {
		log.Info().Str("key", key).Msg("Key")
		if !strings.HasPrefix(key, prefix) {
			log.Info().Str("key", key).Msg("Key does not have prefix")
			continue
		}

		if strings.HasSuffix(key, ".buffer") {
			log.Info().Str("key", key).Msg("Key has suffix")
			continue
		}

		log.Info().Str("key", key).Msg("Key has prefix and suffix")

		// Fetch the page data for each key
		entry, err := kv.Get(key)

		log.Info().Str("key", key).Msg("Key")

		var b types.Button

		log.Info().Interface("entry", entry).Msg("Entry")
		if err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to get value for key")
			continue
		}

		err = json.Unmarshal(entry.Value(), &b)

		if err != nil {
			return nil, err
		}

		buttons = append(buttons, b)

	}

	return buttons, nil
}

func DeleteButton(instanceID string, deviceID string, profileID string, pageID string, buttonID string) error {
	_, kv := natsconn.GetNATSConn()

	// Delete the button
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s", instanceID, deviceID, profileID, pageID, buttonID)
	log.Info().Str("key", key).Msg("Deleting button")
	err := kv.Delete(key)

	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to delete button")
		return err
	}

	// Delete the buffer
	key = fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s.buffer", instanceID, deviceID, profileID, pageID, buttonID)
	log.Info().Str("key", key).Msg("Deleting button buffer")
	err = kv.Delete(key)

	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to delete button buffer")
		return err
	}

	return nil
}

// func updateSettings(buttonKey string, settings Settings) (err error) {
// 	_, kv := natsconn.GetNATSConn()

// 	log.Printf("Setting settings to: %v", settings)
// 	data := Settings{}

// 	// Serialize the Profile struct to JSON
// 	json, err := json.Marshal(data)

// 	if err != nil {
// 		log.Printf("Failed to serialize settings data: %v", err)
// 		return err
// 	}

// 	_, err = kv.Create(buttonKey+".settings", json)

// 	if err != nil {
// 		if err == nats.ErrKeyExists {
// 			log.Printf("Button settings key already exists: %s", buttonKey)
// 		} else {
// 			log.Printf("Failed to create key in KV store: %s %v", buttonKey, err)
// 		}
// 		return err
// 	}

// 	log.Printf("Updated settings id: %v", buttonKey)

// 	return nil
// }

// func updateStateId(buttonKey string, id int) (err error) {
// 	_, kv := natsconn.GetNATSConn()
// 	log.Printf("Setting state ID to: %v", id)
// 	stateId := StateId{
// 		Id: id,
// 	}

// 	// Serialize the Profile struct to JSON
// 	data, err := json.Marshal(stateId)

// 	if err != nil {
// 		log.Printf("Failed to serialize StateId data: %v", err)
// 		return err
// 	}

// 	_, err = kv.Create(buttonKey+".state", data)

// 	if err != nil {
// 		if err == nats.ErrKeyExists {
// 			log.Printf("Button state id key already exists: %s", buttonKey)
// 		} else {
// 			log.Printf("Failed to create key in KV store: %s %v", buttonKey, err)
// 		}
// 		return err
// 	}

// 	log.Printf("Updated state id: %v", buttonKey)

// 	return nil
// }

// func updateTitle(buttonKey string, title string) (err error) {
// 	_, kv := natsconn.GetNATSConn()
// 	log.Printf("Setting title to: %+v", title)
// 	// Serialize the title string to JSON
// 	json, err := json.Marshal(title)

// 	if err != nil {
// 		log.Printf("Failed to serialize title: %+v", err)
// 		return err
// 	}

// 	_, err = kv.Create(buttonKey+".title", json)

// 	if err != nil {
// 		if err == nats.ErrKeyExists {
// 			log.Printf("Button title key already exists: %s", buttonKey)
// 		} else {
// 			log.Printf("Failed to create key in KV store: %s %v", buttonKey, err)
// 		}
// 		return err
// 	}

// 	log.Printf("Updated title: %+v", buttonKey)

// 	return nil
// }

// CreateButton
func CreateButton(instanceID string, deviceID string, profileID string, pageID string, buttonID string) error {
	_, kv := natsconn.GetNATSConn()
	log.Info().Str("instanceID", instanceID).Str("deviceID", deviceID).Str("profileID", profileID).Str("pageID", pageID).Str("buttonID", buttonID).Msg("Creating button")

	// Define the key for the current button
	key := "instances." + instanceID + ".devices." + deviceID + ".profiles." + profileID + ".pages." + pageID + ".buttons." + buttonID

	//var assetPath = env.Get("ASSET_PATH", "")

	// Define a new Button.
	button := types.Button{
		ID:   buttonID,
		UUID: "none",
		States: []types.State{
			{
				ID:        "0",
				ImagePath: env.Get("ASSET_PATH", "") + "images/correct.png",
			},
		},
		State: "0",
		Title: "",
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(button)

	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to marshal button")
		return err
	}

	// Put the serialized data into the KV store
	_, err = kv.Create(key, data)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Error().Err(err).Str("key", key).Msg("Button key already exists")
		} else {
			log.Error().Err(err).Str("key", key).Msg("Failed to create key in KV store")
		}
		return err
	}

	// updateStateId(key, 0)
	// updateSettings(key, Settings{})
	// updateTitle(key, "")
	updateImageBuffer(key, button.States[0].ImagePath)

	log.Info().Str("key", key).Msg("Created button")
	return nil
}

func updateImageBuffer(key string, imagePath string) (err error) {
	log.Info().Str("image_path", imagePath).Str("key", key).Msg("Updating image buffer")
	_, kv := natsconn.GetNATSConn()

	buf, _ := bimg.Read(imagePath)

	_, err = kv.Put(key+".buffer", buf)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Error().Err(err).Str("key", key).Msg("Button buffer key already exists")
		} else {
			log.Error().Err(err).Str("key", key).Msg("Failed to create key in KV store")
		}
		return err
	}

	log.Info().Str("key", key).Msg("Updated button buffer")

	return nil

}
