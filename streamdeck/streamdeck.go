package streamdeck

import (
	"sd/watchers"
	"sd/xl"

	"github.com/karalabe/hid"
)

const ElgatoVendorID = 0x0fd9

type streamdeckType struct {
	Name      string
	ProductID uint16
}

var StreamDeckTypes = []streamdeckType{
	{
		Name:      "Stream Deck XL",
		ProductID: 0x006c,
	},
	// {
	// 	Name: "Pedal",
	// },
	// {
	// 	Name: "Plus",
	// },
}

type StreamDeck struct {
	instanceID string
	device     *hid.Device
}

func New(instanceID string, device *hid.Device) StreamDeck {
	return StreamDeck{
		instanceID: instanceID,
		device:     device,
	}
}

func (sd StreamDeck) Init() {
	go watchers.WatchForButtonChanges()

	if sd.device.Product == "Stream Deck XL" {
		xl := xl.New(sd.instanceID, sd.device)
		xl.Init()
	}
}

// func (sd StreamDeck) GetSerial() string {
// 	return sd.serial
// }

// func (sd StreamDeck) GetProduct() string {
// 	return sd.product
// }

// func (sd StreamDeck) GetProductID() uint16 {
// 	return sd.productID
// }
