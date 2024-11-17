package subscribers

import (
	"sd/stream_decks/xl/utils"

	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

func SubscribeSdInitialize(nc *nats.Conn, device *hid.Device) {
    nc.Subscribe("sd.initialize", func(m *nats.Msg) {
		utils.SetKey(device, 0, "./assets/images/black.jpg")
		utils.SetKey(device, 1, "./assets/images/black.jpg")
		utils.SetKey(device, 2, "./assets/images/black.jpg")
		utils.SetKey(device, 3, "./assets/images/black.jpg")
		utils.SetKey(device, 4, "./assets/images/black.jpg")
		utils.SetKey(device, 5, "./assets/images/black.jpg")
		utils.SetKey(device, 6, "./assets/images/black.jpg")
		utils.SetKey(device, 7, "./assets/images/black.jpg")
		utils.SetKey(device, 8, "./assets/images/black.jpg")
		utils.SetKey(device, 9, "./assets/images/black.jpg")
		utils.SetKey(device, 10, "./assets/images/black.jpg")
		utils.SetKey(device, 11, "./assets/images/black.jpg")
		utils.SetKey(device, 12, "./assets/images/black.jpg")
		utils.SetKey(device, 13, "./assets/images/black.jpg")
		utils.SetKey(device, 14, "./assets/images/black.jpg")
		utils.SetKey(device, 15, "./assets/images/black.jpg")
		utils.SetKey(device, 16, "./assets/images/black.jpg")
		utils.SetKey(device, 17, "./assets/images/black.jpg")
		utils.SetKey(device, 18, "./assets/images/black.jpg")
		utils.SetKey(device, 19, "./assets/images/black.jpg")
		utils.SetKey(device, 20, "./assets/images/black.jpg")
		utils.SetKey(device, 21, "./assets/images/black.jpg")
		utils.SetKey(device, 22, "./assets/images/black.jpg")
		utils.SetKey(device, 23, "./assets/images/black.jpg")
		utils.SetKey(device, 24, "./assets/images/black.jpg")
		utils.SetKey(device, 25, "./assets/images/black.jpg")
		utils.SetKey(device, 26, "./assets/images/black.jpg")
		utils.SetKey(device, 27, "./assets/images/black.jpg")
		utils.SetKey(device, 28, "./assets/images/black.jpg")
		utils.SetKey(device, 29, "./assets/images/black.jpg")
		utils.SetKey(device, 30, "./assets/images/black.jpg")
		utils.SetKey(device, 31, "./assets/images/black.jpg")
	})
}

