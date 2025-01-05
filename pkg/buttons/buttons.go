package buttons

import (
	"encoding/json"
	"sd/pkg/natsconn"

	"github.com/h2non/bimg"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Settings struct {
	URL     string `json:"url,omitempty"`
	Text    string `json:"text,omitempty"`
	Command string `json:"command,omitempty"`
}

// type Title struct {
// }

type State struct {
	Id        string `json:"id"`
	ImagePath string `json:"imagePath"`
}

type StateId struct {
	Id int `json:"id"`
}

type Button struct {
	UUID     string   `json:"uuid"`
	Settings Settings `json:"settings"`
	States   []State  `json:"states"`
	State    string   `json:"state"`
	Title    string   `json:"title"`
}

func GetButton(key string) (button Button, err error) {
	_, kv := natsconn.GetNATSConn()

	entry, err := kv.Get(key)
	if err != nil {
		log.Error().Str("key", key).Err(err).Msg("Failed to retrieve key")
		return Button{}, err
	}

	err = json.Unmarshal(entry.Value(), &button)

	if err != nil {
		log.Error().Err(err).Msg("Failed to parse JSON")
		return Button{}, err
	}

	return button, nil
}

func updateImageBuffer(key string, imagePath string) (err error) {
	_, kv := natsconn.GetNATSConn()

	log.Info().Str("image_path", imagePath).Str("key", key).Msg("Updating image buffer")

	buf, _ := bimg.Read(imagePath)

	_, err = kv.Put(key+".buffer", buf)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Button buffer key already exists: %s", key)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", key, err)
		}
		return err
	}

	log.Info().Str("key", key).Msg("Updated button buffer")

	return nil

}

func updateSettings(buttonKey string, settings Settings) (err error) {
	_, kv := natsconn.GetNATSConn()

	log.Printf("Setting settings to: %v", settings)
	data := Settings{}

	// Serialize the Profile struct to JSON
	json, err := json.Marshal(data)

	if err != nil {
		log.Printf("Failed to serialize settings data: %v", err)
		return err
	}

	_, err = kv.Create(buttonKey+".settings", json)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Button settings key already exists: %s", buttonKey)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", buttonKey, err)
		}
		return err
	}

	log.Printf("Updated settings id: %v", buttonKey)

	return nil
}

func updateStateId(buttonKey string, id int) (err error) {
	_, kv := natsconn.GetNATSConn()
	log.Printf("Setting state ID to: %v", id)
	stateId := StateId{
		Id: id,
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(stateId)

	if err != nil {
		log.Printf("Failed to serialize StateId data: %v", err)
		return err
	}

	_, err = kv.Create(buttonKey+".state", data)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Button state id key already exists: %s", buttonKey)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", buttonKey, err)
		}
		return err
	}

	log.Printf("Updated state id: %v", buttonKey)

	return nil
}

func updateTitle(buttonKey string, title string) (err error) {
	_, kv := natsconn.GetNATSConn()
	log.Printf("Setting title to: %+v", title)
	// Serialize the title string to JSON
	json, err := json.Marshal(title)

	if err != nil {
		log.Printf("Failed to serialize title: %+v", err)
		return err
	}

	_, err = kv.Create(buttonKey+".title", json)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Button title key already exists: %s", buttonKey)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", buttonKey, err)
		}
		return err
	}

	log.Printf("Updated title: %+v", buttonKey)

	return nil
}

// CreateButton
func CreateButton(instanceId string, device *hid.Device, profileId string, pageId string, buttonId string) (err error) {
	_, kv := natsconn.GetNATSConn()
	log.Printf("Creating button: %v", buttonId)

	// Define the key for the current button
	key := "instances." + instanceId + ".devices." + device.Serial + ".profiles." + profileId + ".pages." + pageId + ".buttons." + buttonId

	// Define a new Button.
	button := Button{
		UUID: "",
		States: []State{
			{
				Id:        "0",
				ImagePath: "/home/donovan/.config/sd/buttons/game.png",
			},
		},
		State: "0",
		Title: "",
	}

	// Serialize the Profile struct to JSON
	data, err := json.Marshal(button)

	if err != nil {
		log.Printf("Failed to serialize button data: %v", err)
		return err
	}

	// Put the serialized data into the KV store
	_, err = kv.Create(key, data)

	if err != nil {
		if err == nats.ErrKeyExists {
			log.Printf("Button key already exists: %s", key)
		} else {
			log.Printf("Failed to create key in KV store: %s %v", key, err)
		}
		return err
	}

	updateStateId(key, 0)
	updateSettings(key, Settings{})
	updateTitle(key, "")
	updateImageBuffer(key, button.States[0].ImagePath)

	log.Printf("Created button: %v", buttonId)

	return nil
}

func NewButton() Button {
	return Button{
		States: []State{},
		State:  "0",
		Title:  "",
	}
}
