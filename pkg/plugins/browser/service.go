package browser

import (
	"encoding/json"
	"fmt"
	"sd/pkg/actions"
	"sd/pkg/natsconn"

	"github.com/nats-io/nats.go"
	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
)

type BrowserPlugin struct{}

type OpenURLConfig struct {
	URL string `json:"url"`
}

func (b *BrowserPlugin) Name() string {
	return "browser"
}

func (b *BrowserPlugin) Init() {
	log.Info().Msg("Browser plugin initialized")
	b.openSubscriber()
}

func (b *BrowserPlugin) GetActionTypes() []actions.ActionType {
	return []actions.ActionType{
		"open_url",
	}
}

func (b *BrowserPlugin) ValidateConfig(actionType actions.ActionType, config json.RawMessage) error {
	var cfg OpenURLConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return err
	}
	if cfg.URL == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	return nil
}

func (b *BrowserPlugin) openSubscriber() {
	nc, _ := natsconn.GetNATSConn()
	nc.Subscribe("sd.plugin.browser.open_url", func(msg *nats.Msg) {
		log.Info().Interface("msg", msg.Data).Msg("Browser plugin received message")
		var action actions.ActionInstance
		if err := json.Unmarshal(msg.Data, &action); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal action")
			return
		}

		if url, ok := action.Settings.(map[string]interface{})["url"].(string); ok {
			browser.OpenURL(url)
		}
	})
}
