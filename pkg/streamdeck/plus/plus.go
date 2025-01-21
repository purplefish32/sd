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
	"time"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Plus struct {
	instanceID       string
	device           *hid.Device
	currentProfile   string
	currentPage      string
	wasDialPressed   [4]bool
	wasScreenPressed bool
	lastX            int
	touchScreen      *TouchScreenManager
}

var ProductID uint16 = 0x0084

const VendorID uint16 = 0x0fd9

const (
	DialTurningFlag          = 0x01
	DialTurnRight            = 0x01
	DialTurnLeft             = 0xFF
	DialPressedFlag          = 0x01
	TouchScreenFlag          = 0x01
	TouchPressedFlag         = 0x02
	ButtonPressedFlag        = 0x02
	ScreenWidth              = 800
	ScreenHeight             = 100
	SegmentWidth             = 200 // Each segment is 200px wide (800/4)
	TouchScreenReportLength  = 1024
	TouchScreenPayloadLength = 1008 // 1024 - 16 (header)
	TouchScreenHeaderLength  = 16
	ChunkDelay               = 20 * time.Millisecond
)

type DialEvent struct {
	DialIndex int // 0-3 for dials A-D
	IsTurning bool
	IsPressed bool
	Direction int // 1 for right, -1 for left, 0 for press/release
}

type TouchEvent struct {
	X         int    // X coordinate
	Y         int    // Y coordinate
	IsPressed bool   // Whether the screen is being touched
	Action    string // "tap", "swipe_left", or "swipe_right"
}

func New(instanceID string, device *hid.Device) Plus {
	plus := Plus{
		instanceID: instanceID,
		device:     device,
	}
	plus.touchScreen = NewTouchScreenManager(&plus)
	return plus
}

func (plus *Plus) Init() error {
	// Add reconnection attempt if device is nil
	if plus.device == nil {
		devices := hid.Enumerate(VendorID, ProductID)
		if len(devices) == 0 {
			return fmt.Errorf("no Stream Deck Plus devices found")
		}

		device, err := devices[0].Open()
		if err != nil {
			return fmt.Errorf("failed to open Stream Deck Plus: %w", err)
		}
		plus.device = device
	}

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
			return err
		}

		log.Info().Str("profileId", profile.ID).Msg("Default profile created")

		// Set the profile as the current profile.
		profiles.SetCurrentProfile(plus.instanceID, plus.device.Serial, profile.ID)
		currentProfile = profile
	}

	if currentProfile == nil {
		log.Error().Msg("Failed to get or create current profile")
		return fmt.Errorf("failed to get or create current profile")
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
			return err
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

	// Initialize touch screen with current profile
	if err := plus.touchScreen.UpdateFromProfile(currentProfile); err != nil {
		log.Error().Err(err).Msg("Failed to initialize touch screen")
		return err
	}

	// Watch for profile changes
	go plus.touchScreen.WatchProfileChanges(plus.instanceID)

	// Listen for incoming device input.
	for {
		n, _ := plus.device.Read(buf)
		if n > 0 {
			plus.handleEvent(buf)
		}
	}
}

func BlankKey(device *hid.Device, keyId int, buffer []byte) {
	// Update Key.
	util.SetKeyFromBuffer(device, keyId, buffer)
}

func BlankAllKeys(device *hid.Device) {
	var assetPath = env.Get("ASSET_PATH", "")
	var buffer, err = util.ConvertButtonImageToBuffer(assetPath + "images/black.png")

	if err != nil {
		log.Error().Err(err).Msg("Could not convert blank image to buffer")
	}

	for i := 1; i <= 8; i++ {
		BlankKey(device, i, buffer)
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
			buffer, _ := util.ConvertButtonImageToBuffer(env.Get("ASSET_PATH", "") + "images/black.png")
			BlankKey(device, id, buffer)
		case nats.KeyValuePut:
			var button buttons.Button
			if err := json.Unmarshal(update.Value(), &button); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal button")
				continue
			}
			if len(button.States) > 0 {
				buf, err := util.ConvertButtonImageToBuffer(button.States[0].ImagePath)
				if err != nil {
					log.Error().Err(err).Msg("Failed to create button buffer")
					continue
				}
				BlankKey(device, id, buf)
			}
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

func (plus *Plus) handleTouchEvent(buf []byte) {
	isPressed := buf[4] == TouchPressedFlag
	x := int(buf[5]) | int(buf[6])<<8

	event := TouchEvent{
		X:         x,
		Y:         int(buf[7]),
		IsPressed: isPressed,
	}

	// Track press state for swipe detection
	if isPressed && !plus.wasScreenPressed {
		plus.wasScreenPressed = true
		plus.lastX = x
	} else if !isPressed && plus.wasScreenPressed {
		plus.wasScreenPressed = false
		// Add swipe detection logic here if needed
	}

	plus.handleTouchAction(event)
}

func (plus *Plus) handleTouchAction(event TouchEvent) {
	// Get NATS connection
	nc, _ := natsconn.GetNATSConn()

	// Create a payload for the touch event
	data, err := json.Marshal(event)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal touch event")
		return
	}

	// Publish to NATS with a touch-specific topic
	topic := fmt.Sprintf("instances.%s.devices.%s.touch",
		plus.instanceID, plus.device.Serial)
	nc.Publish(topic, data)
}

func (plus *Plus) handleDialAction(event DialEvent) {
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

func (plus *Plus) handleEvent(buf []byte) {
	// Check for dial events first (they have a specific pattern)
	if buf[0] == 0x01 && buf[1] == 0x03 && buf[2] == 0x05 {
		plus.handleDialEvent(buf)
		return
	}

	// Check for touch events
	if buf[0] == 0x01 && buf[1] == 0x02 && buf[2] == 0x0E {
		plus.handleTouchEvent(buf) // Remove the slice, pass full buffer
		return
	}

	// Handle button events (they start with 0x01)
	if buf[0] == 0x01 {
		plus.handleButtonPress(int(buf[1]))
		return
	}
}

func (plus *Plus) handleButtonEvent(buf []byte) {
	pressedButtons := util.ParseEventBuffer(buf)
	for _, buttonIndex := range pressedButtons {
		// Ignore button up event for now.
		if buttonIndex == 0 {
			continue
		}
		plus.handleButtonPress(buttonIndex)
	}
}

// SetScreenImage sets a full image (800x100) on the touch screen
func (plus *Plus) SetScreenImage(buffer []byte) error {
	remainingBytes := len(buffer)
	iteration := 0

	for remainingBytes > 0 {
		chunkSize := min(remainingBytes, TouchScreenPayloadLength)
		bytesSent := iteration * TouchScreenPayloadLength

		header := []byte{
			0x02, 0x0C, // Fixed header
			0x00, 0x00, // X offset (0 for full screen)
			0x00, 0x00, // Reserved
			0x20, 0x03, // Width (800 in little endian)
			0x64, 0x00, // Height (100 in little endian)
			boolToByte(remainingBytes == chunkSize),      // Final packet flag
			byte(iteration),                              // Page number
			0x00,                                         // Reserved
			byte(chunkSize & 0xFF), byte(chunkSize >> 8), // Payload length (little endian)
			0x00, // Reserved
		}

		chunk := buffer[bytesSent : bytesSent+chunkSize]
		payload := append(header, chunk...)

		err := writeChunkWithDelay(plus.device, payload)
		if err != nil {
			return fmt.Errorf("failed to write chunk %d: %w", iteration, err)
		}

		remainingBytes -= chunkSize
		iteration++
	}

	return nil
}

// SetScreenSegment sets an image on one of the four screen segments (200x100 each)
func (plus *Plus) SetScreenSegment(segment int, buffer []byte) error {
	if segment < 1 || segment > 4 {
		return fmt.Errorf("invalid segment number: %d (must be 1-4)", segment)
	}

	offset := (segment - 1) * SegmentWidth
	offsetBytes := []byte{byte(offset & 0xFF), byte(offset >> 8)}

	remainingBytes := len(buffer)
	iteration := 0

	for remainingBytes > 0 {
		chunkSize := min(remainingBytes, TouchScreenPayloadLength)
		bytesSent := iteration * TouchScreenPayloadLength

		header := []byte{
			0x02, 0x0C, // Fixed header
			offsetBytes[0], offsetBytes[1], // X offset in little endian
			0x00, 0x00, // Reserved
			0xC8, 0x00, // Width (200 in little endian)
			0x64, 0x00, // Height (100 in little endian)
			boolToByte(remainingBytes == chunkSize),      // Final packet flag
			byte(iteration),                              // Page number
			0x00,                                         // Reserved
			byte(chunkSize & 0xFF), byte(chunkSize >> 8), // Payload length (little endian)
			0x00, // Reserved
		}

		chunk := buffer[bytesSent : bytesSent+chunkSize]
		payload := append(header, chunk...)

		err := writeChunkWithDelay(plus.device, payload)
		if err != nil {
			return fmt.Errorf("failed to write chunk %d: %w", iteration, err)
		}

		remainingBytes -= chunkSize
		iteration++
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func writeChunkWithDelay(device *hid.Device, payload []byte) error {
	if device == nil {
		return fmt.Errorf("device is nil")
	}

	// Add padding to match report length
	if len(payload) < TouchScreenReportLength {
		padding := make([]byte, TouchScreenReportLength-len(payload))
		payload = append(payload, padding...)
	}

	_, err := device.Write(payload)
	if err != nil {
		return err
	}
	time.Sleep(ChunkDelay)
	return nil
}
