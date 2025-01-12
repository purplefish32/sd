package main

import (
	"os"
	"sd/cmd/web/server"
	"sd/pkg/env"
	"sd/pkg/util"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Set global time format for logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Str("app", "web").
		Logger()

	log.Info().Msg("Starting application")

	root, err := util.GetProjectRoot()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get project root")
		return
	}

	env.LoadEnv(root + "/cmd/web/.env")

	// Create and start the server
	server := server.NewServer()
	defer server.Close()

	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
