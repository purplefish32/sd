package browser

import (
	"encoding/json"

	natsconn "sd/nats"
	"sd/streamdeck/xl"

	b "github.com/pkg/browser"
	"github.com/rs/zerolog/log"

	"github.com/nats-io/nats.go"
)

// Subscribe sets up the NATS subscription for this plugin.
func OpenSubscriber() {
	nc, _ := natsconn.GetNATSConn()

	nc.Subscribe("sd.plugin.browser.open", func(m *nats.Msg) {
		log.Debug().Msg("Received message for browser plugin")

		var actionInstance xl.ActionInstance

		// Parse the incoming message
		if err := json.Unmarshal(m.Data, &actionInstance); err != nil {
			log.Error().Err(err).Msg("Error unmarshaling JSON")
			return
		}

		// Extract URL directly from Settings
		url, ok := actionInstance.Settings["url"].(string)

		if !ok || url == "" {
			log.Error().Msg("Invalid or empty URL in settings")
			return
		}

		// Open the URL in the default browser
		if err := b.OpenURL(url); err != nil {
			log.Error().Err(err).Msg("Cannot open URL")
			return
		}

		// Log the successful URL opening
		log.Info().Str("URL", url).Msg("Opened URL successfully")
	})
}