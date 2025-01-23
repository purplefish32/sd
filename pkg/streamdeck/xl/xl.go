package xl

import (
	"context"
	"encoding/json"
	"fmt"
	"sd/pkg/env"
	"sd/pkg/natsconn"
	"sd/pkg/store"
	"sd/pkg/types"
	"sd/pkg/util"
	"strconv"
	"strings"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type XL struct {
	instanceID string
	device     *hid.Device
	vendorID   uint16
	productID  uint16
	cancel     context.CancelFunc
	ctx        context.Context
}

func New(instanceID string, device *hid.Device) XL {
	ctx, cancel := context.WithCancel(context.Background())
	return XL{
		instanceID: instanceID,
		device:     device,
		vendorID:   0x0fd9,
		productID:  0x006c,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (xl *XL) Cleanup() {
	if xl.cancel != nil {
		xl.cancel()
	}
	if xl.device != nil {
		xl.device.Close()
	}
}

func (xl *XL) Init() error {
	log.Info().Interface("device", xl.device).Msg("Initializing Stream Deck XL")
	// Add reconnection attempt if device is nil
	if xl.device == nil {
		devices := hid.Enumerate(xl.vendorID, xl.productID)
		if len(devices) == 0 {
			return fmt.Errorf("no Stream Deck XL devices found")
		}

		device, err := devices[0].Open()
		if err != nil {
			return fmt.Errorf("failed to open Stream Deck XL: %w", err)
		}
		xl.device = device
	}

	// Blank all keys.
	xl.blankAllKeys()

	// If there is no profile, create one.
	device := store.GetDevice(xl.instanceID, xl.device.Serial)

	if device.CurrentProfile == "" {
		log.Warn().Msg("No default profile found, creating new profile")
		// Create a new profile.
		profile, _ := store.CreateProfile(xl.instanceID, xl.device.Serial, "Default")

		// Set the profile as the current profile.
		store.SetCurrentProfile(xl.instanceID, xl.device.Serial, profile.ID)

		// Create a new page.
		page, _ := store.CreatePage(xl.instanceID, xl.device.Serial, profile.ID)

		// Set the page as the current page.
		store.SetCurrentPage(xl.instanceID, xl.device.Serial, profile.ID, page.ID)

		// Create 32 blank buttons for the page.
		for i := 0; i < 32; i++ {
			store.CreateButton(xl.instanceID, xl.device.Serial, profile.ID, page.ID, strconv.Itoa(i+1))
		}
	}

	// Start watchers with context
	go xl.watchForButtonChanges(xl.ctx)
	go xl.watchKVForButtonImageBufferChanges(xl.ctx)

	// Start button input loop with context
	go xl.handleButtonInput(xl.ctx)

	return nil
}

// Split out the button input handling into its own function
func (xl *XL) handleButtonInput(ctx context.Context) {
	buf := make([]byte, 512)
	nc, kv := natsconn.GetNATSConn()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := xl.device.Read(buf)
			if err != nil {
				log.Error().Err(err).Msg("Error reading from device")
				return
			}

			if n > 0 {
				pressedButtons := util.ParseEventBuffer(buf)

				// TODO implement long press.
				for _, buttonIndex := range pressedButtons {

					// Ignore button up event for now.
					if buttonIndex == 0 {
						continue
					}

					log.Info().Int("buttonIndex", buttonIndex).Msg("Button pressed")

					currentProfile := store.GetCurrentProfile(xl.instanceID, xl.device.Serial)

					currentPage := store.GetCurrentPage(xl.instanceID, xl.device.Serial, currentProfile.ID)

					key := "instances." + xl.instanceID + ".devices." + xl.device.Serial + ".profiles." + currentProfile.ID + ".pages." + currentPage.ID + ".buttons." + strconv.Itoa(buttonIndex)

					// Get the associated data from the NATS KV Store.
					entry, _ := nats.KeyValue.Get(kv, key)

					// Unmarshal the JSON into the Payload struct
					var payload types.ActionInstance

					if err := json.Unmarshal(entry.Value(), &payload); err != nil {
						log.Error().Err(err).Msg("Failed to unmarshal JSON from KV store")
						return
					}

					// Use the `UUID` field as the topic
					if payload.UUID == "" {
						log.Error().Msg("Missing UUID field in JSON payload")
						return
					}

					// Publish Action Instance to NATS.
					nc.Publish(payload.UUID, entry.Value())
				}
			}
		}
	}
}

func (xl *XL) blankKey(keyId int) {
	var assetPath = env.Get("ASSET_PATH", "")
	var buffer, err = util.ConvertImageToBuffer(assetPath+"images/black.png", 96)

	if err != nil {
		log.Error().Err(err).Msg("Could not convert blank image to buffer")
	}
	// Update Key with rotation for XL
	util.SetKeyFromBuffer(xl.device, keyId, buffer, true)
}

func (xl *XL) blankAllKeys() {
	for i := 1; i <= 32; i++ {
		xl.blankKey(i)
	}
}

func (xl *XL) watchForButtonChanges(ctx context.Context) {
	_, kv := natsconn.GetNATSConn()

	buttonPattern := "instances.*.devices." + xl.device.Serial + ".profiles.*.pages.*.buttons.*"

	watcher, err := kv.Watch(buttonPattern)
	if err != nil {
		log.Warn().Err(err).Msg("Error creating watcher")
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-watcher.Updates():
			if update == nil {
				continue
			}

			// Get button number from the key
			segments := strings.Split(update.Key(), ".")
			buttonNum := segments[len(segments)-1]

			id, err := strconv.Atoi(buttonNum)
			if err != nil {
				continue
			}

			switch update.Operation() {
			case nats.KeyValueDelete:
				xl.blankKey(id)
			case nats.KeyValuePut:
				var button types.Button
				if err := json.Unmarshal(update.Value(), &button); err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal button")
					continue
				}
				if len(button.States) > 0 {
					buf, err := util.ConvertImageToBuffer(button.States[0].ImagePath, 96)
					if err != nil {
						log.Error().Err(err).Msg("Failed to create button buffer")
						continue
					}

					key := update.Key() + ".buffer"
					kv.Put(key, buf)
				}
			}
		}
	}
}

func (xl *XL) watchKVForButtonImageBufferChanges(ctx context.Context) {
	_, kv := natsconn.GetNATSConn()

	// Get current profile and page, with error checking
	currentProfile := store.GetCurrentProfile(xl.instanceID, xl.device.Serial)

	if currentProfile == nil {
		log.Warn().Msg("No current profile found")
		return
	}

	currentPage := store.GetCurrentPage(xl.instanceID, xl.device.Serial, currentProfile.ID)

	if currentPage == nil {
		log.Warn().Msg("No current page found")
		return
	}

	// Create watcher with error handling
	pattern := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.*.buffer",
		xl.instanceID, xl.device.Serial, currentProfile.ID, currentPage.ID)

	watcher, err := kv.Watch(pattern)
	if err != nil {
		log.Error().Err(err).Msg("Error creating watcher")
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-watcher.Updates():
			if update == nil {
				continue
			}

			switch update.Operation() {
			case nats.KeyValuePut:
				log.Info().Str("key", update.Key()).Msg("Key added/updated")
				// Get Stream Deck key id from the kv key.

				// Split the string by the delimiter ".".
				segments := strings.Split(update.Key(), ".")

				// Get the last segment.
				sdKeyId := segments[len(segments)-2]

				// Convert to an int.
				id, err := strconv.Atoi(sdKeyId)

				if err != nil {
					// ... handle error.
					panic(err)
				}

				// Update Key with rotation for XL
				util.SetKeyFromBuffer(xl.device, id, update.Value(), true)
			case nats.KeyValueDelete:
				log.Info().Str("key", update.Key()).Msg("Key deleted")
			default:
				log.Info().Str("key", update.Key()).Msg("Unknown operation")
			}
		}
	}
}
