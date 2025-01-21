package streamdeck

import (
	"sd/pkg/streamdeck/xl"
	"sync"

	"fmt"

	"github.com/karalabe/hid"
)

const (
	VendorIDElgato = 0x0fd9
	ProductIDXL    = 0x006c
	ProductIDPlus  = 0x0084
	ProductIDPedal = 0x0086
)

var devices = struct {
	sync.RWMutex
	list map[string]*StreamDeck
}{
	list: make(map[string]*StreamDeck),
}

type StreamDeck struct {
	instanceID string
	device     *hid.Device
}

func New(instanceID string, deviceID string, productID uint16) error {
	devices := hid.Enumerate(VendorIDElgato, productID)
	if len(devices) == 0 {
		return fmt.Errorf("no devices found with product ID: %x", productID)
	}

	device, err := devices[0].Open()
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}

	switch productID {
	case ProductIDXL:
		xlDevice := xl.New(instanceID, device)
		return xlDevice.Init()
	// case ProductIDPlus:
	// 	plusDevice := plus.New(instanceID, device)
	// 	return plusDevice.Init()
	// case ProductIDPedal:
	// 	pedalDevice := pedal.New(instanceID, device)
	// 	return pedalDevice.Init()
	default:
		return fmt.Errorf("unsupported device type: %x", productID)
	}
}

func (sd StreamDeck) Init() {
	if sd.device.Product == "Stream Deck XL" {
		xl := xl.New(sd.instanceID, sd.device)
		xl.Init()
	}

	// if sd.device.Product == "Stream Deck Plus" {
	// 	plus := plus.New(sd.instanceID, sd.device)
	// 	plus.Init()
	// }

	// if sd.device.Product == "Stream Deck Pedal" {
	// 	pedal := pedal.New(sd.instanceID, sd.device)
	// 	pedal.Init()
	// }
}

func RemoveDevice(deviceID string) {
	devices.Lock()
	defer devices.Unlock()
	if device, exists := devices.list[deviceID]; exists {
		if device.device != nil {
			device.device.Close()
		}
		delete(devices.list, deviceID)
	}
}
