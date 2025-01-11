package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Set global time format for logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	log.Info().Msg("Starting application")

	// Load the .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, using default configuration")
	}

	// Create and start the server
	server := NewServer()
	defer server.Close()

	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
