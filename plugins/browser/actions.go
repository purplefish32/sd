package browser

import (
	"encoding/json"
	"sd/actions"
	"sd/natsconn"

	b "github.com/pkg/browser"
	"github.com/rs/zerolog/log"

	"github.com/nats-io/nats.go"
)

type Settings struct {
	URL string `json:"url"`
}

// Subscribe sets up the NATS subscription for this plugin.
func OpenSubscriber() {
	nc, _ := natsconn.GetNATSConn()

	if _, err := nc.Subscribe("sd.plugin.browser.open", func(m *nats.Msg) {
		log.Debug().Msg("Received message for browser plugin")

		var actionInstance actions.ActionInstance

		// Parse the incoming message
		if err := json.Unmarshal(m.Data, &actionInstance); err != nil {
			log.Error().Err(err).Msg("Error unmarshaling JSON")
			return
		}

		// Convert actionInstance.Settings to Settings
		settingsMap, ok := actionInstance.Settings.(map[string]any)
		if !ok {
			log.Error().Msg("Settings is not a valid object")
			return
		}

		settingsBytes, err := json.Marshal(settingsMap)
		if err != nil {
			log.Error().Err(err).Msg("Error marshaling settings to JSON")
			return
		}

		var settings Settings
		if err := json.Unmarshal(settingsBytes, &settings); err != nil {
			log.Error().Err(err).Msg("Error unmarshaling settings to Settings")
			return
		}

		// Validate URL
		if settings.URL == "" {
			log.Error().Msg("URL is empty")
			return
		}

		// Open the URL in the default browser
		if err := b.OpenURL(settings.URL); err != nil {
			log.Error().Err(err).Msg("Cannot open URL")
			return
		}

		// Log the successful URL opening
		log.Info().Str("URL", settings.URL).Msg("Opened URL successfully")
	}); err != nil {
		log.Fatal().Err(err).Msg("Failed to subscribe to sd.plugin.browser.open")
	}
}
