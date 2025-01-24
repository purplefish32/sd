package plus

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
	"time"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// Constants for device configuration
const (
	vendorID  = 0x0fd9
	productID = 0x0084
	numKeys   = 8
	keySize   = 120

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

type Plus struct {
	instanceID       string
	device           *hid.Device
	cancel           context.CancelFunc
	ctx              context.Context
	wasDialPressed   [4]bool
	wasScreenPressed bool
	lastX            int
	touchScreen      *TouchScreenManager
}

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
	ctx, cancel := context.WithCancel(context.Background())
	plus := Plus{
		instanceID: instanceID,
		device:     device,
		ctx:        ctx,
		cancel:     cancel,
	}
	plus.touchScreen = NewTouchScreenManager(&plus)
	return plus
}

func (plus *Plus) Cleanup() {
	if plus.cancel != nil {
		plus.cancel()
	}
	if plus.device != nil {
		plus.device.Close()
	}
}

func (plus *Plus) Init() error {
	log.Info().Interface("device", plus.device).Msg("Initializing Stream Deck Plus")

	if err := plus.ensureDeviceConnection(); err != nil {
		return err
	}

	plus.blankAllKeys()

	if err := plus.ensureDefaultProfile(); err != nil {
		return err
	}

	// Start watchers and input handlers
	go plus.watchForButtonChanges(plus.ctx)
	go plus.watchKVForButtonImageBufferChanges(plus.ctx)
	go plus.handleInput(plus.ctx)

	// Initialize touch screen with current profile
	currentProfile := store.GetCurrentProfile(plus.instanceID, plus.device.Serial)
	if !currentProfile.IsEmpty() {
		if err := plus.touchScreen.UpdateFromProfile(&currentProfile); err != nil {
			log.Error().Err(err).Msg("Failed to initialize touch screen")
		}
	}

	return nil
}

func BlankKey(device *hid.Device, keyId int, buffer []byte) {
	// Update Key.
	util.SetKeyFromBuffer(device, keyId, buffer, false)
}

func (plus *Plus) blankAllKeys() {
	var assetPath = env.Get("ASSET_PATH", "")
	var buffer, err = util.ConvertButtonImageToBuffer(assetPath+"images/correct.png", keySize)

	if err != nil {
		log.Error().Err(err).Msg("Could not convert blank image to buffer")
	}

	for i := 1; i <= numKeys; i++ {
		BlankKey(plus.device, i, buffer)
	}
}

func (plus *Plus) watchKVForButtonImageBufferChanges(ctx context.Context) {
	_, kv := natsconn.GetNATSConn()

	currentProfile := store.GetCurrentProfile(plus.instanceID, plus.device.Serial)
	currentPage := store.GetCurrentPage(plus.instanceID, plus.device.Serial, currentProfile.ID)

	pattern := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.*.buffer",
		plus.instanceID, plus.device.Serial, currentProfile.ID, currentPage.ID)

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
				segments := strings.Split(update.Key(), ".")
				sdKeyId := segments[len(segments)-2]
				id, err := strconv.Atoi(sdKeyId)
				if err != nil {
					continue
				}
				util.SetKeyFromBuffer(plus.device, id, update.Value(), false)
			}
		}
	}
}

func (plus *Plus) watchForButtonChanges(ctx context.Context) {
	_, kv := natsconn.GetNATSConn()

	buttonPattern := fmt.Sprintf("instances.*.devices.%s.profiles.*.pages.*.buttons.*", plus.device.Serial)
	watcher, err := kv.Watch(buttonPattern)
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

			segments := strings.Split(update.Key(), ".")
			buttonNum := segments[len(segments)-1]

			id, err := strconv.Atoi(buttonNum)
			if err != nil {
				continue
			}

			switch update.Operation() {
			case nats.KeyValueDelete:
				buffer, _ := util.ConvertButtonImageToBuffer(env.Get("ASSET_PATH", "")+"images/correct.png", keySize)
				BlankKey(plus.device, id, buffer)
			case nats.KeyValuePut:
				var button types.Button
				if err := json.Unmarshal(update.Value(), &button); err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal button")
					continue
				}
				if len(button.States) > 0 {
					buf, err := util.ConvertButtonImageToBuffer(button.States[0].ImagePath, keySize)
					if err != nil {
						log.Error().Err(err).Msg("Failed to create button buffer")
						continue
					}
					BlankKey(plus.device, id, buf)
				}
			}
		}
	}
}

func (plus *Plus) handleButtonPress(buttonIndex int) {
	currentProfile := store.GetCurrentProfile(plus.instanceID, plus.device.Serial)
	if currentProfile.IsEmpty() {
		return
	}

	currentPage := store.GetCurrentPage(plus.instanceID, plus.device.Serial, currentProfile.ID)
	if currentPage.IsEmpty() {
		return
	}

	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s",
		plus.instanceID, plus.device.Serial, currentProfile.ID, currentPage.ID, strconv.Itoa(buttonIndex))

	button, err := store.GetButton(key)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to get button configuration")
		return
	}

	// Get NATS connection
	nc, _ := natsconn.GetNATSConn()

	// Create ActionInstance from Button
	actionInstance := types.Button{
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

func (plus *Plus) handleInput(ctx context.Context) {
	buf := make([]byte, 512)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := plus.device.Read(buf)
			if err != nil {
				log.Error().Err(err).Msg("Error reading from device")
				return
			}

			if n > 0 {
				// Check for dial events first
				if buf[0] == 0x01 && buf[1] == 0x03 && buf[2] == 0x05 {
					plus.handleDialEvent(buf)
					continue
				}

				// Check for touch events
				if buf[0] == 0x01 && buf[1] == 0x02 && buf[2] == 0x0E {
					plus.handleTouchEvent(buf)
					continue
				}

				// Handle button events
				if buf[0] == 0x01 {
					plus.handleButtonPress(int(buf[1]))
				}
			}
		}
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

func (plus *Plus) ensureDeviceConnection() error {
	if plus.device != nil {
		return nil
	}

	devices := hid.Enumerate(vendorID, productID)
	if len(devices) == 0 {
		return fmt.Errorf("no Stream Deck Plus devices found")
	}

	device, err := devices[0].Open()
	if err != nil {
		return fmt.Errorf("failed to open Stream Deck Plus: %w", err)
	}
	plus.device = device
	return nil
}

func (plus *Plus) ensureDefaultProfile() error {
	device := store.GetDevice(plus.instanceID, plus.device.Serial)
	if device.CurrentProfile != "" {
		return nil
	}

	profile, err := store.CreateProfile(plus.instanceID, plus.device.Serial, "Default")
	if err != nil {
		return fmt.Errorf("failed to create default profile: %w", err)
	}

	store.SetCurrentProfile(plus.instanceID, plus.device.Serial, profile.ID)

	page, err := store.CreatePage(plus.instanceID, plus.device.Serial, profile.ID)
	if err != nil {
		return fmt.Errorf("failed to create default page: %w", err)
	}

	store.SetCurrentPage(plus.instanceID, plus.device.Serial, profile.ID, page.ID)

	// Create blank buttons
	for i := 0; i < numKeys; i++ {
		if err := store.CreateButton(plus.instanceID, plus.device.Serial, profile.ID, page.ID, strconv.Itoa(i+1)); err != nil {
			return fmt.Errorf("failed to create button %d: %w", i+1, err)
		}
	}
	return nil
}
