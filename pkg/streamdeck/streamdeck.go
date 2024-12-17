package streamdeck

import (
	"sd/pkg/streamdeck/pedal"
	"sd/pkg/streamdeck/xl"

	"github.com/karalabe/hid"
	"github.com/rs/zerolog/log"
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
	{
		Name:      "Stream Deck Pedal",
		ProductID: 0x0086,
	},
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
	log.Warn().Msg(sd.device.Product)

	if sd.device.Product == "Stream Deck XL" {
		xl := xl.New(sd.instanceID, sd.device)
		xl.Init()
	}

	if sd.device.Product == "Stream Deck Pedal" {
		pedal := pedal.New(sd.instanceID, sd.device)
		pedal.Init()
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
