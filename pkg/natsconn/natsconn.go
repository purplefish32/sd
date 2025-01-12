package natsconn

import (
	"os"
	"sd/pkg/env"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

var (
	nc   *nats.Conn
	kv   nats.KeyValue
	once sync.Once
)

// GetNATSConn returns a singleton NATS connection.
func GetNATSConn() (*nats.Conn, nats.KeyValue) {
	once.Do(func() {
		// Get NATS server address from the .env file.
		natsUrl := env.Get("NATS_URL", "nats://127.0.0.1:4222")

		log.Info().Str("NATS_URL", natsUrl).Msg("Connecting to NATS server")

		if natsUrl == "" {
			log.Fatal().Msg("NATS_URL is not set in the .env file")
		}

		natsKVBucket := env.Get("NATS_KV_BUCKET", "sd")

		log.Info().Str("NATS_KV_BUCKET", natsKVBucket).Msg("NATS_KV_BUCKET")

		if natsUrl == "" {
			log.Fatal().Msg("NATS_KV_BUCKET is not set in the .env file")
		}

		// Add connection options
		opts := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(-1),
			nats.ReconnectWait(2 * time.Second),
			nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
				log.Warn().Err(err).Msg("NATS disconnected")
			}),
			nats.ReconnectHandler(func(nc *nats.Conn) {
				log.Info().Msg("NATS reconnected")
			}),
		}

		// Connect to NATS server with retry options
		var err error
		nc, err = nats.Connect(natsUrl, opts...)
		if err != nil {
			log.Fatal().Err(err).Str("NATS_URL", natsUrl).Msg("Failed connecting to NATS server")
			os.Exit(1)
		}

		// Enable JetStream Context
		js, err := nc.JetStream()

		if err != nil {
			log.Fatal().Err(err).Msg("Error enabling JetStream")
			os.Exit(1) // Explicitly terminate the program
		}

		// Check if the Key-Value bucket already exists
		kv, err = js.KeyValue(natsKVBucket)

		// Try to access the bucket
		if err == nats.ErrBucketNotFound {
			// Create the bucket if it doesn't exist
			kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
				Bucket: natsKVBucket, // Name of the bucket
			})

			if err != nil {
				log.Fatal().Err(err).Str("bucket", natsKVBucket).Msg("Error creating Key-Value bucket")
				os.Exit(1) // Explicitly terminate the program
			}
		} else if err != nil {
			log.Fatal().Err(err).Msg("Error accessing Key-Value bucket")
			os.Exit(1) // Explicitly terminate the program
		}
	})

	return nc, kv
}
