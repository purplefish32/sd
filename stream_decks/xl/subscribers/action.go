package subscribers

import (
	"encoding/json"
	"log"
	"os/exec"

	"github.com/go-vgo/robotgo"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
	"github.com/pkg/browser"
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


func SubscribeSdAction(nc *nats.Conn, device *hid.Device) {
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