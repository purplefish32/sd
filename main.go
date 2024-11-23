package main

import (
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"sd/core"
	"sd/plugins/browser"
	"sd/plugins/command"
	"sd/plugins/keyboard"
	"sd/streamdeck"
	streamdeckPedal "sd/streamdeck/pedal"
	streamdeckXl "sd/streamdeck/xl"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/karalabe/hid"
	"github.com/nats-io/nats.go"
)

func main() {
	// Retrieve or create the instance UUID
	instanceID := getOrCreateUUID()

	// Print the instance UUID
	log.Printf("Instance UUID: %s\n", instanceID)

	// Load the .env file
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get NATS server address from the .env file
	natsServer := os.Getenv("NATS_SERVER")

	if natsServer == "" {
		log.Fatal("NATS_SERVER is not set in the .env file")
	}

	// Connect to NATS
	nc, err := nats.Connect(natsServer)

	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}

    defer nc.Close()

	// Enable JetStream Context
	js, err := nc.JetStream()

	if err != nil {
		log.Fatalf("Error enabling JetStream: %v", err)
	}

	// Check if the Key-Value bucket already exists
	kv, err := js.KeyValue("sd") // TODO get this from environment.
	
	// Try to access the bucket
	if err == nats.ErrBucketNotFound {
		// Create the bucket if it doesn't exist
		kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket: "sd", // Name of the bucket
		})
		if err != nil {
			log.Fatalf("Error creating Key-Value bucket: %v", err)
		}
		log.Println("Key-Value bucket 'sd' created successfully")
	} else if err != nil {
		log.Fatalf("Error accessing Key-Value bucket: %v", err)
	} else {
		log.Println("Key-Value bucket 'sd' already exists")
	}

	// Register plugins
	registry := core.NewPluginRegistry()
	registry.Register(&browser.BrowserPlugin{})
	registry.Register(&command.CommandPlugin{})
	registry.Register(&keyboard.KeyboardPlugin{})

	// Subscribe plugins to NATS topics
	for _, plugin := range registry.All() {
		log.Printf("Registering plugin: %s", plugin.Name())
		if err := plugin.Subscribe(nc); err != nil {
			log.Printf("Error subscribing plugin %s: %v", plugin.Name(), err)
		} else {
			log.Printf("Plugin %s subscribed successfully.", plugin.Name())
		}
	}

	// Define the devices you want to manage
	deviceTypes := []struct {
		Name      string
		VendorID  uint16
		ProductID uint16
		Initialize func(nc *nats.Conn, instanceID string, device *hid.Device, kv nats.KeyValue)
		RequiresKV bool

	}{
		{"Stream Deck XL", streamdeck.VendorID, streamdeckXl.ProductID, streamdeckXl.Initialize, true},
		{"Stream Deck Pedal", streamdeck.VendorID, streamdeckPedal.ProductID, streamdeckPedal.Initialize, false},
		//{"Stream Deck +", streamdeck.VendorID, streamdeckPlus.ProductID, streamdeckPlus.Initialize},
	}

	// Process each device type
	for _, deviceType := range deviceTypes {
		go func(dt struct {
			Name       string
			VendorID   uint16
			ProductID  uint16
			Initialize func(nc *nats.Conn, instanceID string, device *hid.Device, kv nats.KeyValue)
			RequiresKV bool
		}) {
			// Find the devices
			hidDevices := hid.Enumerate(dt.VendorID, dt.ProductID)
			if len(hidDevices) == 0 {
				log.Printf("No %s found", dt.Name)
				return
			}

			// Process each device of this type
			for i, hidDeviceInfo := range hidDevices {
				go func(deviceIndex int, deviceInfo hid.DeviceInfo) {
					// Open the device
					hidDevice, err := deviceInfo.Open()
					if err != nil {
						log.Printf("Failed to open %s (Device %d): %v", dt.Name, deviceIndex, err)
						return
					}

					// Initialize the device
					log.Printf("Initializing %s (Device %d)...", dt.Name, deviceIndex)


					// Conditionally pass kv or nil
					if dt.RequiresKV {
						dt.Initialize(nc, instanceID, hidDevice, kv)  // Pass pointer to kv if required
					} else {
						dt.Initialize(nc, instanceID, hidDevice, nil)  // Pass nil if not required
					}

					log.Printf("%s (Device %d) initialized successfully.", dt.Name, deviceIndex)
				}(i, hidDeviceInfo)
			}
		}(deviceType)
	}

	// iconBuilderSubscribers.CreateIconBuffer(nc)

	// Keep the main program running
	select {} // Blocks forever
}

func getOrCreateUUID() string {
	// Use a directory in the user's home folder
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error retrieving user home directory: %v", err)
	}
	uuidDir := filepath.Join(homeDir, ".config/sd")
	uuidFilePath := filepath.Join(uuidDir, "instance-id")

	// Ensure the directory exists
	if _, err := os.Stat(uuidDir); os.IsNotExist(err) {
		err := os.MkdirAll(uuidDir, 0755) // Create the directory
		if err != nil {
			log.Fatalf("Error creating directory %s: %v", uuidDir, err)
		}
	}

	// Check if the UUID file exists
	if _, err := os.Stat(uuidFilePath); err == nil {
		// Read the existing UUID
		data, err := os.ReadFile(uuidFilePath)
		if err != nil {
			log.Fatalf("Error reading UUID file: %v", err)
		}
		return string(data)
	}

	// Generate a new UUID
	id := uuid.New()
	idStr := id.String()

	// Save the UUID to the file
	err = os.WriteFile(uuidFilePath, []byte(idStr), 0600)
	if err != nil {
		log.Fatalf("Error saving UUID to file: %v", err)
	}

	return idStr
}