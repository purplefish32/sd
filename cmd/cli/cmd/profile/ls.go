package profile

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
	Short: "List all profile IDs",
	Run: func(cmd *cobra.Command, args []string) {
		// Load the current configuration

		// Connect to NATS.
		_, kv := natsconn.GetNATSConn()

		entries, err := kv.Keys()

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to list keys")
		}

		currentProfileId := viper.GetString("current-profile")

		// Extract unique instance IDs
		profileIDs := make(map[string]struct{}) // Use a map to ensure uniqueness
		for _, key := range entries {
			parts := strings.Split(key, ".")

			if len(parts) > 4 && parts[4] == "profiles" && parts[5] != "current" { // TODO change this, it is hacky.

				profileIDs[parts[5]] = struct{}{}
			}
		}

		// Create table using the helper function
		table := util.NewTable(cmd.OutOrStdout())
		table.SetHeader([]string{"PROFILE ID", "CURRENT"})

		// Add rows
		for id := range profileIDs {
			currentMark := ""
			if id == currentProfileId {
				currentMark = "CURRENT"
			}
			table.Append([]string{id, currentMark})
		}

		// Render the table
		table.Render()
	},
}

func init() {
	profileCmd.AddCommand(lsCmd)
}
