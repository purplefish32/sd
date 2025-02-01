package store

import (
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"sd/pkg/types"
	"strings"

	"github.com/rs/zerolog/log"
)

func GetDevice(instanceID string, deviceID string) *types.Device {
	_, kv := natsconn.GetNATSConn()

	key := fmt.Sprintf("instances.%s.devices.%s", instanceID, deviceID)
	log.Info().Str("key", key).Msg("Key")

	entry, err := kv.Get(key)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get device")
		return nil
	}

	var device types.Device

	if err := json.Unmarshal(entry.Value(), &device); err != nil {
		log.Error().Err(err).Msg("Unmarshal error")
		return nil
	}

	return &device
}

func GetDevices(instanceID string) []types.Device {
	_, kv := natsconn.GetNATSConn()

	keyList, err := kv.ListKeys()

	if err != nil {
		log.Warn().Err(err).Msg("Failed to get devices")
		return nil
	}

	devices := make([]types.Device, 0)

	for key := range keyList.Keys() {
		if !strings.Contains(key, fmt.Sprintf("instances.%s.devices.", instanceID)) {
			continue
		}

		parts := strings.Split(key, ".")

		if len(parts) != 4 {
			continue
		}

		entry, err := kv.Get(key)

		if err != nil {
			log.Warn().Err(err).Str("key", key).Msg("Skipping invalid device entry")
			continue
		}

		var device types.Device

		if err := json.Unmarshal(entry.Value(), &device); err != nil {
			log.Warn().Err(err).Str("key", key).Msg("Skipping malformed device data")
			continue
		}

		devices = append(devices, types.Device{
			ID:             device.ID,
			Instance:       device.Instance,
			Type:           device.Type,
			Status:         device.Status,
			CurrentProfile: device.CurrentProfile,
		})

	}

	return devices
}

func UpdateDevice(instanceID string, device *types.Device) (*types.Device, error) {
	if instanceID == "" || device == nil {
		return nil, fmt.Errorf("instanceID, and device are required")
	}
	_, kv := natsconn.GetNATSConn()

	key := fmt.Sprintf("instances.%s.devices.%s", instanceID, device.ID)

	json, err := json.Marshal(device)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal device: %w", err)
	}

	_, err = kv.Put(key, json)
	if err != nil {
		return nil, fmt.Errorf("failed to save device: %w", err)
	}

	return device, nil
}
