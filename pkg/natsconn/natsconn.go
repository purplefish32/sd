package natsconn

import (
	"os"
	"sd/pkg/env"
	"sync"

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
		env.LoadEnv()

		// Get NATS server address from the .env file.
		natsServer := os.Getenv("NATS_SERVER")

		if natsServer == "" {
			log.Fatal().Msg("NATS_SERVER is not set in the .env file")
		}

		natsKVBucket := os.Getenv("NATS_KV_BUCKET")

		if natsServer == "" {
			log.Fatal().Msg("NATS_KV_BUCKET is not set in the .env file")
		}

		var err error

		// Connect to NATS server.
		nc, err = nats.Connect(natsServer)

		if err != nil {
			log.Fatal().Str("nats_server", natsServer).Msg("Failed connecting to NATS server")
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
