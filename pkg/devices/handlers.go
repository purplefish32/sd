package devices

import (
	"encoding/json"
	"fmt"
	"time"

	"sd/pkg/natsconn"
)

func HandleConnect(instanceID string, deviceID string, productID uint16) error {
	_, kv := natsconn.GetNATSConn()

	deviceType := determineDeviceType(productID)
	if deviceType == "unknown" {
		return fmt.Errorf("unknown device type for product ID: %x", productID)
	}

	key := fmt.Sprintf("instances.%s.devices.%s", instanceID, deviceID)
	info := Info{
		Type:      deviceType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    "connected",
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	_, err = kv.Put(key, data)
	return err
}

func HandleDisconnect(instanceID string, deviceID string) error {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s", instanceID, deviceID)

	entry, err := kv.Get(key)
	if err != nil {
		return err
	}

	var info Info
	if err := json.Unmarshal(entry.Value(), &info); err != nil {
		return err
	}

	info.Status = "disconnected"
	info.UpdatedAt = time.Now()

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	_, err = kv.Put(key, data)
	return err
}

func determineDeviceType(productID uint16) string {
	switch productID {
	case ProductIDXL:
		return TypeXL
	case ProductIDPlus:
		return TypePlus
	case ProductIDPedal:
		return TypePedal
	default:
		return "unknown"
	}
}
