package xl

import (
	"encoding/json"
	"fmt"
	"sd/pkg/actions"
	"sd/pkg/buttons"
	"sd/pkg/env"
	"sd/pkg/natsconn"
	"sd/pkg/pages"
	"sd/pkg/profiles"
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
}

var ProductID uint16 = 0x006c

const VendorID uint16 = 0x0fd9

func New(instanceID string, device *hid.Device) XL {
	return XL{
		instanceID: instanceID,
		device:     device,
	}
}

func (xl *XL) Init() error {
	// Add reconnection attempt if device is nil
	if xl.device == nil {
		devices := hid.Enumerate(VendorID, ProductID)
		if len(devices) == 0 {
			return fmt.Errorf("no Stream Deck XL devices found")
		}

		device, err := devices[0].Open()
		if err != nil {
			return fmt.Errorf("failed to open Stream Deck XL: %w", err)
		}
		xl.device = device
	}

	log.Info().
		Str("device_serial", xl.device.Serial).
		Msg("Stream Deck XL Initialization")

	// Blank all keys.
	BlankAllKeys(xl.device)

	currentProfile := profiles.GetCurrentProfile(xl.instanceID, xl.device.Serial)

	// If no default profile exists, create one and set is as the default profile.
	if currentProfile == nil {
		log.Warn().Msg("Current profile not found creating one")

		// Create a new profile.
		profile, _ := profiles.CreateProfile(xl.instanceID, xl.device.Serial, "Default")

		log.Info().Str("profileId", profile.ID).Msg("Profile created")

		// Set the profile as the current profile.
		profiles.SetCurrentProfile(xl.instanceID, xl.device.Serial, profile.ID)
	}

	currentProfile = profiles.GetCurrentProfile(xl.instanceID, xl.device.Serial)

	log.Info().Interface("current_profile", currentProfile).Msg("Current profile")

	currentPage := pages.GetCurrentPage(xl.instanceID, xl.device.Serial, currentProfile.ID)

	// If no default page exists, create one and set is as the default page for the given profile.
	if currentPage == nil {
		log.Warn().Msg("Current page not found creating one")

		// Create a new page.
		page, err := pages.CreatePage(xl.instanceID, xl.device.Serial, currentProfile.ID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create page")
			return nil
		}

		log.Info().Interface("page", page).Msg("Page created")

		// Set the page as the current page.
		pages.SetCurrentPage(xl.instanceID, xl.device.Serial, currentProfile.ID, page.ID)
	}

	// Buffer for outgoing events.
	buf := make([]byte, 512)

	// Get NATS connection an KV store.
	nc, kv := natsconn.GetNATSConn()

	go WatchForButtonChanges(xl.device)

	// Listen for incoming device input.
	for {
		n, _ := xl.device.Read(buf)

		if n > 0 {
			pressedButtons := util.ParseEventBuffer(buf)

			// TODO implement long press.
			for _, buttonIndex := range pressedButtons {

				// Ignore button up event for now.
				if buttonIndex == 0 {
					continue
				}

				key := "instances." + xl.instanceID + ".devices." + xl.device.Serial + ".profiles." + currentProfile.ID + ".pages." + currentPage.ID + ".buttons." + strconv.Itoa(buttonIndex)

				// Get the associated data from the NATS KV Store.
				entry, _ := nats.KeyValue.Get(kv, key)

				// if err != nil {
				// 	log.Warn().Err(err).Msg("Failed to get value from KV store")
				// 	continue
				// }

				// Unmarshal the JSON into the Payload struct
				var payload actions.ActionInstance

				if err := json.Unmarshal(entry.Value(), &payload); err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal JSON from KV store")
					return nil
				}

				// Use the `UUID` field as the topic
				if payload.UUID == "" {
					log.Error().Msg("Missing UUID field in JSON payload")
					return nil
				}

				// Publish Action Instance to NATS.
				nc.Publish(payload.UUID, entry.Value())
			}
		}
	}
}

func BlankKey(device *hid.Device, keyId int, buffer []byte) {
	// Update Key.
	util.SetKeyFromBuffer(device, keyId, buffer)
}

func BlankAllKeys(device *hid.Device) {
	var assetPath = env.Get("ASSET_PATH", "")
	var buffer, err = util.ConvertImageToRotatedBuffer(assetPath+"images/black.png", 96)

	if err != nil {
		log.Error().Err(err).Msg("Could not convert blank image to buffer")
	}

	for i := 1; i <= 32; i++ {
		BlankKey(device, i, buffer)
	}
}

func WatchForButtonChanges(device *hid.Device) {
	_, kv := natsconn.GetNATSConn()

	buttonPattern := fmt.Sprintf("instances.*.devices.%s.profiles.*.pages.*.buttons.*", device.Serial)
	watcher, err := kv.Watch(buttonPattern)
	if err != nil {
		log.Error().Err(err).Msg("Error creating watcher")
		return
	}
	defer watcher.Stop()

	for update := range watcher.Updates() {
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
			// Blank the key when button is deleted
			buffer, _ := util.ConvertImageToRotatedBuffer(env.Get("ASSET_PATH", "")+"images/black.png", 96)
			BlankKey(device, id, buffer)
		case nats.KeyValuePut:
			var button buttons.Button
			if err := json.Unmarshal(update.Value(), &button); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal button")
				continue
			}
			if len(button.States) > 0 {
				buf, err := util.ConvertImageToRotatedBuffer(button.States[0].ImagePath, 96)
				if err != nil {
					log.Error().Err(err).Msg("Failed to create button buffer")
					continue
				}
				BlankKey(device, id, buf)
			}
		}
	}
}
