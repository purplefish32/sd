package keyboard

import (
	"encoding/json"
	"log"

	"github.com/go-vgo/robotgo"
	"github.com/nats-io/nats.go"
)

// KeyboardPlugin represents the keyboard plugin
type KeyboardPlugin struct{}

// Name returns the name of the plugin
func (k *KeyboardPlugin) Name() string {
	return "keyboard"
}

// Subscribe sets up the NATS subscription for this plugin
func (k *KeyboardPlugin) Subscribe(nc *nats.Conn) error {
	_, err := nc.Subscribe("sd.plugin.keyboard", func(m *nats.Msg) {
		log.Println("Keyboard Plugin received NATS message:", string(m.Data))

		var msg struct {
			Data struct {
				Action string `json:"action"`
				Text   string `json:"text"`
			} `json:"data"`
		}

		// Parse the incoming message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			log.Printf("Error unmarshaling JSON: %v\n", err)
			return
		}

		// Perform the desired action
		if msg.Data.Action == "type" {
			log.Println("Typing text:", msg.Data.Text)
			robotgo.TypeStr(msg.Data.Text)
		}
	})
	return err
}