package pedal

import (
	"context"
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"sd/pkg/store"
	"sd/pkg/types"
	"sd/pkg/util"
	"strconv"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// Constants for device configuration
const (
	vendorID  = 0x0fd9
	productID = 0x0086
	numKeys   = 3
	keySize   = 96
)

type Pedal struct {
	instanceID string
	device     *hid.Device
	cancel     context.CancelFunc
	ctx        context.Context
}

func New(instanceID string, device *hid.Device) Pedal {
	ctx, cancel := context.WithCancel(context.Background())
	return Pedal{
		instanceID: instanceID,
		device:     device,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (pedal *Pedal) Cleanup() {
	if pedal.cancel != nil {
		pedal.cancel()
	}
	if pedal.device != nil {
		pedal.device.Close()
	}
}

func (pedal *Pedal) Init() error {
	log.Info().Interface("device", pedal.device).Msg("Initializing Stream Deck Pedal")

	if err := pedal.ensureDeviceConnection(); err != nil {
		return err
	}

	if err := pedal.ensureDefaultProfile(); err != nil {
		return err
	}

	// Start input handler
	go pedal.handleButtonInput(pedal.ctx)

	return nil
}

func (pedal *Pedal) ensureDeviceConnection() error {
	if pedal.device != nil {
		return nil
	}

	devices := hid.Enumerate(vendorID, productID)
	if len(devices) == 0 {
		return fmt.Errorf("no Stream Deck Pedal devices found")
	}

	device, err := devices[0].Open()
	if err != nil {
		return fmt.Errorf("failed to open Stream Deck Pedal: %w", err)
	}
	pedal.device = device
	return nil
}

func (pedal *Pedal) ensureDefaultProfile() error {
	device := store.GetDevice(pedal.instanceID, pedal.device.Serial)
	if device.CurrentProfile != "" {
		return nil
	}

	profile, err := store.CreateProfile(pedal.instanceID, pedal.device.Serial, "Default")
	if err != nil {
		return fmt.Errorf("failed to create default profile: %w", err)
	}

	store.SetCurrentProfile(pedal.instanceID, pedal.device.Serial, profile.ID)

	page, err := store.CreatePage(pedal.instanceID, pedal.device.Serial, profile.ID)
	if err != nil {
		return fmt.Errorf("failed to create default page: %w", err)
	}

	store.SetCurrentPage(pedal.instanceID, pedal.device.Serial, profile.ID, page.ID)

	// Create blank switches for pedal (3 switches)
	for i := 0; i < numKeys; i++ {
		if err := store.CreateButton(pedal.instanceID, pedal.device.Serial, profile.ID, page.ID, strconv.Itoa(i+1)); err != nil {
			return fmt.Errorf("failed to create switch %d: %w", i+1, err)
		}
	}
	return nil
}

func (pedal *Pedal) handleButtonPress(buttonIndex int, nc *nats.Conn, kv nats.KeyValue) error {
	currentProfile := store.GetCurrentProfile(pedal.instanceID, pedal.device.Serial)
	if currentProfile.IsEmpty() {
		return fmt.Errorf("no current profile found")
	}

	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.switches.%d",
		pedal.instanceID, pedal.device.Serial, currentProfile.ID, buttonIndex)

	entry, err := kv.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get switch data: %w", err)
	}

	var payload types.ActionInstance
	if err := json.Unmarshal(entry.Value(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal switch data: %w", err)
	}

	if payload.UUID == "" {
		return fmt.Errorf("missing UUID in payload")
	}

	return nc.Publish(payload.UUID, entry.Value())
}

func (pedal *Pedal) handleButtonInput(ctx context.Context) {
	buf := make([]byte, 512)
	nc, kv := natsconn.GetNATSConn()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := pedal.device.Read(buf)
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

					log.Info().Int("buttonIndex", buttonIndex).Msg("Switch pressed")

					if err := pedal.handleButtonPress(buttonIndex, nc, kv); err != nil {
						log.Error().Err(err).Msg("Error handling switch press")
					}
				}
			}
		}
	}
}
