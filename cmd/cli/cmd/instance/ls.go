package instance

import (
	"sd/cmd/cli/pkg/config"
	"sd/pkg/instance"
	"sd/pkg/natsconn"
	"sd/pkg/util"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all instance IDs",
	Run: func(cmd *cobra.Command, args []string) {
		// Load the current configuration
		conf, _ := config.LoadConfig()

		// Connect to NATS.
		_, kv := natsconn.GetNATSConn()

		entries, err := kv.Keys()

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to list keys")
		}

		localInstanceId := instance.GetInstanceId()
		currentInstanceId := conf.CurrentInstance

		// Extract unique instance IDs
		instanceIDs := make(map[string]struct{}) // Use a map to ensure uniqueness
		for _, key := range entries {
			parts := strings.Split(key, ".")
			if len(parts) > 1 && parts[0] == "instances" {
				instanceIDs[parts[1]] = struct{}{}
			}
		}

		// Create table
		// Create table using the helper function
		table := util.NewTable(cmd.OutOrStdout())
		table.SetHeader([]string{"INSTANCE ID", "LOCAL", "CURRENT"})

		// Add rows
		for id := range instanceIDs {
			localMark := "" // Default: no mark
			if id == localInstanceId {
				localMark = "LOCAL" // Mark the local instance with a star
			}
			currentMark := "" // Default: no mark
			if id == currentInstanceId {
				currentMark = "CURRENT" // Mark the local instance with a star
			}
			table.Append([]string{id, localMark, currentMark})
		}

		// Render the table
		table.Render()
	},
}

func init() {
	instanceCmd.AddCommand(lsCmd)
}
