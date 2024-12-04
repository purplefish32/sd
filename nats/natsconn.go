package natsconn

import (
	"os"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// GetNATSConn returns a singleton NATS connection.
func GetNATSConn() (*nats.Conn, nats.KeyValue) {

	var (
		nc   *nats.Conn
		kv   nats.KeyValue
		once sync.Once
	)

	once.Do(func() {
		log.Info().Msg("Init NATS connection")

		// Get NATS server address from the .env file.
		natsServer := os.Getenv("NATS_SERVER")

		if natsServer == "" {
			log.Fatal().Msg("NATS_SERVER is not set in the .env file")
		}

		var err error

		// Connect to NATS server.
		nc, err = nats.Connect(natsServer)

		if err != nil {
			log.Fatal().Str("nats_server", natsServer).Msg("Failed connecting to NATS server")
		}

		log.Info().Str("nats_server", natsServer).Msg("NATS server connection successful")

		// Enable JetStream Context
		js, err := nc.JetStream()

		if err != nil {
			log.Fatal().Err(err).Msg("Error enabling JetStream")
			os.Exit(1) // Explicitly terminate the program
		}

		var bucket = "sd" // TODO get this from environment.

		// Check if the Key-Value bucket already exists
		kv, err = js.KeyValue(bucket)

		// Try to access the bucket
		if err == nats.ErrBucketNotFound {
			// Create the bucket if it doesn't exist
			kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
				Bucket: bucket, // Name of the bucket
			})

			if err != nil {
				log.Fatal().Err(err).Str("bucket", bucket).Msg("Error creating Key-Value bucket")
				os.Exit(1) // Explicitly terminate the program
			}

			log.Info().Str("bucket", bucket).Msg("Key-Value bucket created")
		} else if err != nil {
			log.Fatal().Err(err).Msg("Error accessing Key-Value bucket")
			os.Exit(1) // Explicitly terminate the program
		}

		log.Info().Str("bucket", bucket).Msg("Key-Value bucket exists")
	})

	return nc, kv
}