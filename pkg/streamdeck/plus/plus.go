package plus

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

type Plus struct {
	instanceID     string
	device         *hid.Device
	currentProfile string
	currentPage    string
	wasDialPressed [4]bool
}

var ProductID uint16 = 0x006c

const (
	DialTurningFlag = 0x01
	DialTurnRight   = 0x01
	DialTurnLeft    = 0xFF
	DialPressedFlag = 0x01
)

type DialEvent struct {
	DialIndex int // 0-3 for dials A-D
	IsTurning bool
	IsPressed bool
	Direction int // 1 for right, -1 for left, 0 for press/release
}

func New(instanceID string, device *hid.Device) Plus {
	return Plus{
		instanceID: instanceID,
		device:     device,
	}
}

func (plus Plus) Init() {
	log.Info().
		Str("device_serial", plus.device.Serial).
		Msg("Stream Deck Plus Initialization")

	// Blank all keys.
	BlankAllKeys(plus.device)

	currentProfile := profiles.GetCurrentProfile(plus.instanceID, plus.device.Serial)

	// If no default profile exists, create one and set it as the default profile.
	if currentProfile == nil {
		log.Info().Msg("Creating default profile")

		// Create a new profile.
		profile, err := profiles.CreateProfile(plus.instanceID, plus.device.Serial, "Default")
		if err != nil {
			log.Error().Err(err).Msg("Failed to create default profile")
			return
		}

		log.Info().Str("profileId", profile.ID).Msg("Default profile created")

		// Set the profile as the current profile.
		profiles.SetCurrentProfile(plus.instanceID, plus.device.Serial, profile.ID)
		currentProfile = profile
	}

	if currentProfile == nil {
		log.Error().Msg("Failed to get or create current profile")
		return
	}

	log.Info().Interface("current_profile", currentProfile).Msg("Current profile")

	currentPage := pages.GetCurrentPage(plus.instanceID, plus.device.Serial, currentProfile.ID)

	// If no default page exists, create one and set it as the default page for the given profile.
	if currentPage == nil {
		log.Info().Msg("Creating default page")

		// Create a new page.
		page, err := pages.CreatePage(plus.instanceID, plus.device.Serial, currentProfile.ID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create default page")
			return
		}

		log.Info().Interface("page", page).Msg("Default page created")

		// Set the page as the current page.
		pages.SetCurrentPage(plus.instanceID, plus.device.Serial, currentProfile.ID, page.ID)
		currentPage = &page
	}

	// Buffer for outgoing events.
	buf := make([]byte, 512)

	go WatchForButtonChanges(plus.device)
	go WatchKVForButtonImageBufferChanges(plus.instanceID, plus.device)

	// Listen for incoming device input.
	for {
		n, _ := plus.device.Read(buf)

		if n > 0 {
			// Check if it's a dial event
			if buf[0] == 0x01 && buf[1] == 0x03 && buf[2] == 0x05 {
				plus.handleDialEvent(buf)
				continue
			}

			pressedButtons := util.ParseEventBuffer(buf)

			// TODO implement long press.
			for _, buttonIndex := range pressedButtons {
				// Ignore button up event for now.
				if buttonIndex == 0 {
					log.Debug().Interface("device", plus.device).Int("button_index", buttonIndex).Msg("Button released")
					continue
				}

				plus.handleButtonPress(buttonIndex)
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
	var buffer, err = util.ConvertImageToBuffer(assetPath+"images/black.png", 120)

	if err != nil {
		log.Error().Err(err).Msg("Could not convert blank image to buffer")
	}

	for i := 1; i <= 8; i++ {
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

		// Skip buffer keys
		if strings.HasSuffix(buttonNum, ".buffer") {
			continue
		}

		id, err := strconv.Atoi(buttonNum)
		if err != nil {
			continue
		}

		switch update.Operation() {
		case nats.KeyValueDelete:
			// Blank the key when button is deleted
			buffer, _ := util.ConvertImageToBuffer(env.Get("ASSET_PATH", "")+"images/black.png", 120)
			BlankKey(device, id, buffer)
		case nats.KeyValuePut:
			var button buttons.Button
			if err := json.Unmarshal(update.Value(), &button); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal button")
				continue
			}
			if len(button.States) > 0 {
				buf, err := util.ConvertImageToBuffer(button.States[0].ImagePath, 120)
				if err != nil {
					log.Error().Err(err).Msg("Failed to create button buffer")
					continue
				}
				BlankKey(device, id, buf)
			}
		}
	}
}

func WatchKVForButtonImageBufferChanges(instanceId string, device *hid.Device) {
	// Add contextual information to the logger for this function
	log := log.With().
		Str("instanceId", instanceId).
		Str("deviceSerial", device.Serial).
		Logger()

	_, kv := natsconn.GetNATSConn()

	currentProfile := profiles.GetCurrentProfile(instanceId, device.Serial)
	currentPage := pages.GetCurrentPage(instanceId, device.Serial, currentProfile.ID)

	// Start watching the KV bucket for updates for a specific profile and page.
	watcher, err := kv.Watch("instances." + instanceId + ".devices." + device.Serial + ".profiles." + currentProfile.ID + ".pages." + currentPage.ID + ".buttons.*.buffer")
	defer watcher.Stop()

	if err != nil {
		log.Error().Err(err).Msg("Error creating watcher")
	}

	// Flag to track when all initial values have been processed.
	initialValuesProcessed := false

	// Start the watch loop.
	for update := range watcher.Updates() {
		// If the update is nil, it means all initial values have been received.
		if update == nil {
			if !initialValuesProcessed {
				log.Info().Msg("All initial values have been processed. Waiting for updates")
				initialValuesProcessed = true
			}
			// Continue listening for future updates, so don't break here.
			continue
		}

		// Process the update.
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

			// Update Key.
			util.SetKeyFromBuffer(device, id, update.Value())
		case nats.KeyValueDelete:
			log.Info().Str("key", update.Key()).Msg("Key deleted")
		default:
			log.Info().Str("key", update.Key()).Msg("Unknown operation")
		}
	}
}

func (d *Plus) handleButtonPress(buttonIndex int) {
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
		d.instanceID, d.device.Serial, d.currentProfile, d.currentPage, strconv.Itoa(buttonIndex))

	button, err := buttons.GetButton(key)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to get button configuration")
		return
	}

	// Get NATS connection
	nc, _ := natsconn.GetNATSConn()

	// Create ActionInstance from Button
	actionInstance := actions.ActionInstance{
		UUID:     button.UUID,
		Settings: button.Settings,
		State:    button.State,
		States:   button.States,
		Title:    button.Title,
	}

	// Marshal the action instance
	data, err := json.Marshal(actionInstance)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal action instance")
		return
	}

	// Publish to NATS using the UUID as the topic
	nc.Publish(button.UUID, data)
}

func (plus *Plus) handleDialEvent(buf []byte) {
	isTurning := buf[4] == DialTurningFlag

	// Check each dial (A through D)
	for dialIndex := 0; dialIndex < 4; dialIndex++ {
		dialValue := buf[5+dialIndex]

		// Skip inactive dials
		if dialValue == 0 && !plus.wasDialPressed[dialIndex] && !isTurning {
			continue
		}

		// Create event for each dial
		event := DialEvent{
			DialIndex: dialIndex + 1,
			IsTurning: isTurning,
		}

		if isTurning {
			switch dialValue {
			case DialTurnRight:
				event.Direction = 1
				log.Info().
					Int("dial", event.DialIndex).
					Msg("Dial turned right")
			case DialTurnLeft:
				event.Direction = -1
				log.Info().
					Int("dial", event.DialIndex).
					Msg("Dial turned left")
			default:
				continue
			}
		} else {
			// Not turning - handle press/release
			if dialValue == DialPressedFlag {
				event.IsPressed = true
				plus.wasDialPressed[dialIndex] = true
				log.Info().
					Int("dial", event.DialIndex).
					Msg("Dial pressed")
			} else if dialValue == 0 && plus.wasDialPressed[dialIndex] {
				event.IsPressed = false
				plus.wasDialPressed[dialIndex] = false
				log.Info().
					Int("dial", event.DialIndex).
					Msg("Dial released")
			} else {
				continue
			}
		}

		plus.handleDialAction(event)
	}
}

func (plus Plus) handleDialAction(event DialEvent) {
	// Get NATS connection
	nc, _ := natsconn.GetNATSConn()

	// Create a payload for the dial event
	data, err := json.Marshal(event)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal dial event")
		return
	}

	// Publish to NATS with a dial-specific topic
	topic := fmt.Sprintf("instances.%s.devices.%s.dials.%d",
		plus.instanceID, plus.device.Serial, event.DialIndex)
	nc.Publish(topic, data)
}
