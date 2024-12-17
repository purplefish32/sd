package device

import (
	"sd/pkg/natsconn"
	"sd/pkg/util"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all device IDs",
	Run: func(cmd *cobra.Command, args []string) {
		// Load the current configuration

		// Connect to NATS.
		_, kv := natsconn.GetNATSConn()

		entries, err := kv.Keys()

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to list keys")
		}

		currentDeviceId := viper.GetString("current-device")

		// Extract unique instance IDs
		deviceIDs := make(map[string]struct{}) // Use a map to ensure uniqueness
		for _, key := range entries {
			parts := strings.Split(key, ".")

			if len(parts) > 1 && parts[2] == "devices" {
				deviceIDs[parts[3]] = struct{}{}
			}
		}

		// Create table using the helper function
		table := util.NewTable(cmd.OutOrStdout())
		table.SetHeader([]string{"DEVICE ID", "CURRENT"})

		// Add rows
		for id := range deviceIDs {
			currentMark := ""
			if id == currentDeviceId {
				currentMark = "CURRENT"
			}
			table.Append([]string{id, currentMark})
		}

		// Render the table
		table.Render()
	},
}

func init() {
	deviceCmd.AddCommand(lsCmd)
}
