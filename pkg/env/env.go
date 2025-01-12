package env

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// LoadEnv loads environment variables from a .env file if it exists
func LoadEnv(path string) {
	err := godotenv.Load(path)

	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("No .env file found, falling back to environment variables")
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
