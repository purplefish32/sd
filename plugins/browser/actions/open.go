package actions

import (
	"encoding/json"
	"log"

	natsconn "sd/nats"

	b "github.com/pkg/browser"

	"github.com/nats-io/nats.go"
)

var msg struct {
	Url    string `json:"url"`
}

// Subscribe sets up the NATS subscription for this plugin.
func SubscribeActionOpen(pluginNamespace string) {
	nc, _ := natsconn.GetNATSConn()

	nc.Subscribe(pluginNamespace +".open", func(m *nats.Msg) {

		// Parse the incoming message.
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			log.Printf("Error unmarshaling JSON: %v\n", err)
			return
		}

		// Open the URL in the default browser.
		if err := b.OpenURL(msg.Url); err != nil {
			log.Printf("Error opening URL: %v\n", err)
		}

	})
}