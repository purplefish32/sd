package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Set global time format for logger.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Configure the global logger.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	log.Info().Msg("Starting application")

	// Retrieve or create the instance UUID.
	//instanceID := instance.GetOrCreateInstanceUUID()

	// Load the .env file.
	err := godotenv.Load()

	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
		os.Exit(1) // Explicitly terminate the program.
	}
}
