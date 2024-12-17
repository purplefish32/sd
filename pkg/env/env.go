package env

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// LoadEnv loads environment variables from a .env file if it exists
func LoadEnv() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Error().Err(err).Msg("No .env file found, falling back to environment variables")
	}
}

// Get retrieves the value of an environment variable or a default value if unset
func Get(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
