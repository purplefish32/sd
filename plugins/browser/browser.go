package browser

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/pkg/browser"
)

type pluginData struct {
	Action string `json:"action"`
	Url string `json:"url"`
}

type pluginMessage struct {
	Id string `json:"id"`
	Data pluginData `json:"data"`
}

// BrowserPlugin represents the browser plugin
type BrowserPlugin struct{}

// Name returns the name of the plugin
func (b *BrowserPlugin) Name() string {
	return "browser"
}

// Subscribe sets up the NATS subscription for this plugin
func (b *BrowserPlugin) Subscribe(nc *nats.Conn) error {
	_, err := nc.Subscribe("sd.plugin.browser", func(m *nats.Msg) {
		log.Println("Browser Plugin received NATS message:", string(m.Data))

		var msg struct {
			Data struct {
				Action string `json:"action"`
				Url    string `json:"url"`
			} `json:"data"`
		}

		// Parse the incoming message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			log.Printf("Error unmarshaling JSON: %v\n", err)
			return
		}

		// Perform the desired action
		if msg.Data.Action == "open" {
			if err := browser.OpenURL(msg.Data.Url); err != nil {
				log.Printf("Error opening URL: %v\n", err)
			}
		}
	})
	return err
}

func OpenURL(url string) {
	err := browser.OpenURL(url)

	if err != nil {
		log.Fatal(err)
	}
}
