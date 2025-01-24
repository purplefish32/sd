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

// Constants for device configuration
const (
	vendorID  = 0x0fd9
	productID = 0x006c
	numKeys   = 32
	keySize   = 96
)

type XL struct {
	instanceID string
	device     *hid.Device
	cancel     context.CancelFunc
	ctx        context.Context
}

func New(instanceID string, device *hid.Device) XL {
	ctx, cancel := context.WithCancel(context.Background())
	return XL{
		instanceID: instanceID,
		device:     device,
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

	if err := xl.ensureDeviceConnection(); err != nil {
		return err
	}

	xl.blankAllKeys()

	if err := xl.ensureDefaultProfile(); err != nil {
		return err
	}

	// Start watchers and input handler
	go xl.watchForButtonChanges(xl.ctx)
	go xl.watchKVForButtonImageBufferChanges(xl.ctx)
	go xl.handleButtonInput(xl.ctx)

	return nil
}

func (xl *XL) ensureDeviceConnection() error {
	if xl.device != nil {
		return nil
	}

	devices := hid.Enumerate(vendorID, productID)
	if len(devices) == 0 {
		return fmt.Errorf("no Stream Deck XL devices found")
	}

	device, err := devices[0].Open()
	if err != nil {
		return fmt.Errorf("failed to open Stream Deck XL: %w", err)
	}
	xl.device = device
	return nil
}

func (xl *XL) ensureDefaultProfile() error {
	device := store.GetDevice(xl.instanceID, xl.device.Serial)
	if device.CurrentProfile != "" {
		return nil
	}

	profile, err := store.CreateProfile(xl.instanceID, xl.device.Serial, "Default")
	if err != nil {
		return fmt.Errorf("failed to create default profile: %w", err)
	}

	store.SetCurrentProfile(xl.instanceID, xl.device.Serial, profile.ID)

	page, err := store.CreatePage(xl.instanceID, xl.device.Serial, profile.ID)
	if err != nil {
		return fmt.Errorf("failed to create default page: %w", err)
	}

	store.SetCurrentPage(xl.instanceID, xl.device.Serial, profile.ID, page.ID)

	// Create blank buttons
	for i := 0; i < numKeys; i++ {
		if err := store.CreateButton(xl.instanceID, xl.device.Serial, profile.ID, page.ID, strconv.Itoa(i+1)); err != nil {
			return fmt.Errorf("failed to create button %d: %w", i+1, err)
		}
	}
	return nil
}

func (xl *XL) handleButtonPress(buttonIndex int, nc *nats.Conn, kv nats.KeyValue) error {
	currentProfile := store.GetCurrentProfile(xl.instanceID, xl.device.Serial)
	currentPage := store.GetCurrentPage(xl.instanceID, xl.device.Serial, currentProfile.ID)

	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%d",
		xl.instanceID, xl.device.Serial, currentProfile.ID, currentPage.ID, buttonIndex)

	entry, err := kv.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get button data: %w", err)
	}

	var payload types.ActionInstance
	if err := json.Unmarshal(entry.Value(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal button data: %w", err)
	}

	if payload.UUID == "" {
		return fmt.Errorf("missing UUID in payload")
	}

	return nc.Publish(payload.UUID, entry.Value())
}

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

					if err := xl.handleButtonPress(buttonIndex, nc, kv); err != nil {
						log.Error().Err(err).Msg("Error handling button press")
					}
				}
			}
		}
	}
}

func (xl *XL) blankKey(keyId int) {
	var assetPath = env.Get("ASSET_PATH", "")
	var buffer, err = util.ConvertImageToBuffer(assetPath+"images/correct.png", keySize)

	if err != nil {
		log.Error().Err(err).Msg("Could not convert blank image to buffer")
	}
	// Update Key with rotation for XL
	util.SetKeyFromBuffer(xl.device, keyId, buffer, true)
}

func (xl *XL) blankAllKeys() {
	for i := 1; i <= numKeys; i++ {
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

	if currentProfile.IsEmpty() {
		log.Warn().Msg("No current profile found")
		return
	}

	currentPage := store.GetCurrentPage(xl.instanceID, xl.device.Serial, currentProfile.ID)

	if currentPage.IsEmpty() {
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
