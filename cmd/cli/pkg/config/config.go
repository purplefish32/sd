package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the CLI configuration
type Config struct {
	CurrentInstance string `json:"current_instance"`
}

var configFile = filepath.Join(os.Getenv("HOME"), ".sd", "config.json")

// LoadConfig loads the configuration from the file
func LoadConfig() (*Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil // Return default config if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to the file
func SaveConfig(config *Config) error {
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
