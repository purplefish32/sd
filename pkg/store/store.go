package store

import (
	"encoding/json"
	"fmt"
	"sd/pkg/natsconn"
	"sd/pkg/types"
)

func GetProfile(instanceID, deviceID, profileID string) (*types.Profile, error) {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profileID)

	entry, err := kv.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	var profile types.Profile
	if err := json.Unmarshal(entry.Value(), &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

func UpdateProfile(instanceID, deviceID, profileID string, profile *types.Profile) error {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profileID)

	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	_, err = kv.Put(key, data)
	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

func DeleteProfile(instanceID, deviceID, profileID string) error {
	_, kv := natsconn.GetNATSConn()
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s", instanceID, deviceID, profileID)

	return kv.Delete(key)
}
