package main

import (
	"net/http"
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

	// Get project root
	root, err := util.GetProjectRoot()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get project root")
	}

	env.LoadEnv(root + "/cmd/web/.env")

	// Create and start the server
	server := server.NewServer()
	defer server.Close()

	// Set up static file server with absolute path
	fs := http.FileServer(http.Dir(root + "/cmd/web/assets"))
	r := server.Router()
	r.Handle("/assets/*", http.StripPrefix("/assets/", fs))

	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
