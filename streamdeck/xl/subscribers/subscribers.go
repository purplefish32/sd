package subscribers

import (
	"encoding/json"
	"log"
	"os/exec"

	"github.com/go-vgo/robotgo"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/pkg/browser"

	"sd/streamdeck/xl/utils"
)

type ActionMessage struct {
	Id string `json:"id"`
	Pattern string `json:"pattern"`
	Data ActionMessageData `json:"data"`
}

type ActionMessageData struct {
	Type string `json:"type"`
	Url string `json:"url"`
}

type UpdateMessageData struct {
	Key int `json:"key"`
	Image string `json:"image"`
}

type UpdateMessage struct {
	Id string `json:"id"`
	Pattern string `json:"pattern"`
	Data UpdateMessageData `json:"data"`
}

func SubscribeUpdate(nc *nats.Conn, device *hid.Device) {
	nc.Subscribe("sd.update", func(m *nats.Msg) {

		// Parse the JSON message
		var event UpdateMessage

		err := json.Unmarshal(m.Data, &event)

		if err != nil {
			log.Printf("Failed to parse JSON message: %v", err)
			return
		}

		utils.SetKey(device, event.Data.Key-1, event.Data.Image)
	})
}


func SubscribeAction(nc *nats.Conn, device *hid.Device) {
	nc.Subscribe("sd.action", func(m *nats.Msg) {
		// Unmarshal the JSON payload into the ActionEvent struct
		var actionMessage ActionMessage
		err := json.Unmarshal(m.Data, &actionMessage)
		if err != nil {
			log.Printf("Error unmarshaling JSON: %v\n", err)
			return
		}


		if(actionMessage.Data.Type == "keyboard") {
			log.Println("Action is type Keyboard")
			text := "Hello, World!"

			// Simulate typing text
			robotgo.TypeStr(text)
		}

		if(actionMessage.Data.Type == "browser") {
			err := browser.OpenURL(actionMessage.Data.Url)
			if err != nil {
				log.Fatal(err)
			}
		}

		if(actionMessage.Data.Type == "command") {
			// Define the command and its arguments
			cmd := exec.Command("notify-send", "Hello", "This is a test notification!")

			// Run the command
			err := cmd.Run()
			if err != nil {
				log.Fatalf("Failed to send notification: %v", err)
			}
		}

		if(actionMessage.Data.Type == "radio") {
			// Define the command and its arguments
			cmd := exec.Command("radio")

			// Run the command
			err := cmd.Run()
			if err != nil {
				log.Fatalf("Failed to send notification: %v", err)
			}
		}

		if(actionMessage.Data.Type == "shutdown") {
			// Define the command and its arguments
			cmd := exec.Command("shutdown", "-h", "now")

			// Run the command
			err := cmd.Run()
			if err != nil {
				log.Fatalf("Failed to shut down: %v", err)
			}
		}
	})
}

func SubscribeInitialize(nc *nats.Conn, device *hid.Device) {
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

