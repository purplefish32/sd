package streamdeck

import (
	"sd/pkg/streamdeck/pedal"
	"sd/pkg/streamdeck/plus"
	"sd/pkg/streamdeck/xl"
	"sync"

	"github.com/karalabe/hid"
)

const ElgatoVendorID = 0x0fd9

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

func RemoveDevice(devicePath string) {
	devices.Lock()
	defer devices.Unlock()

	if sd, exists := devices.list[devicePath]; exists {
		// Close the device and clean up resources.
		sd.device.Close()
		delete(devices.list, devicePath)
	}
}

func New(instanceID string, device *hid.Device) StreamDeck {
	return StreamDeck{
		instanceID: instanceID,
		device:     device,
	}
}

func (sd StreamDeck) Init() {
	if sd.device.Product == "Stream Deck XL" {
		xl := xl.New(sd.instanceID, sd.device)
		xl.Init()
	}

	if sd.device.Product == "Stream Deck Plus" {
		plus := plus.New(sd.instanceID, sd.device)
		plus.Init()
	}

	if sd.device.Product == "Stream Deck Pedal" {
		pedal := pedal.New(sd.instanceID, sd.device)
		pedal.Init()
	}
}
