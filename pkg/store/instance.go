package store

import (
	"os"
	"path/filepath"
	"sd/pkg/natsconn"
	"sd/pkg/types"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func GetInstances() ([]types.Instance, error) {
	_, kv := natsconn.GetNATSConn()

	keys, err := kv.Keys()
	if err != nil {
		return nil, err
	}

	instances := make([]types.Instance, 0)

	for _, key := range keys {

		// If the key doesn't start with the prefix, skip it
		if !strings.Contains(key, "instances.") {
			continue
		}

		parts := strings.Split(key, ".")

		if len(parts) != 2 {
			continue
		}

		instanceID := parts[1]

		instances = append(instances, types.Instance{
			ID: instanceID,
		})

	}

	// Add a dummy instance for testing
	// TODO: Remove this once we have a real instance
	instances = append(instances, types.Instance{ID: "db45c7bb-af46-48aa-b2a0-d222586c7909"})

	return instances, nil
}

func GetInstanceId() string {
	// Use a directory in the user's home folder
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving user home directory")
	}

	uuidDir := filepath.Join(homeDir, ".config/sd")
	uuidFilePath := filepath.Join(uuidDir, "instance-id")

	// Read the existing UUID
	data, err := os.ReadFile(uuidFilePath)

	uuid := string(data)

	if err != nil {
		log.Fatal().Err(err).Msg("Error reading UUID file")
		os.Exit(1)
	}

	return uuid
}

func GetOrCreateInstanceUUID() string {

	// Use a directory in the user's home folder
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving user home directory")
	}

	uuidDir := filepath.Join(homeDir, ".config/sd")
	uuidFilePath := filepath.Join(uuidDir, "instance-id")

	// Ensure the directory exists
	if _, err := os.Stat(uuidDir); os.IsNotExist(err) {
		err := os.MkdirAll(uuidDir, 0755) // Create the directory

		if err != nil {
			log.Fatal().Err(err).Str("uuidDir", uuidDir).Msg("Error creating directory")
			os.Exit(1)
		}
	}

	// Check if the UUID file exists
	if _, err := os.Stat(uuidFilePath); err == nil {

		// Read the existing UUID
		data, err := os.ReadFile(uuidFilePath)

		uuid := string(data)

		if err != nil {
			log.Fatal().Err(err).Msg("Error reading UUID file")
			os.Exit(1)
		}

		log.Info().Str("uuid", uuid).Msg("UUID file exists")

		return uuid
	}

	// Generate a new UUID
	id := uuid.New()
	uuid := id.String()

	// Save the UUID to the file
	err = os.WriteFile(uuidFilePath, []byte(uuid), 0600)

	if err != nil {
		log.Fatal().Err(err).Msg("Error saving UUID to file")
	}

	log.Info().Str("uuid", uuid).Msg("UUID file created")

	return uuid
}
