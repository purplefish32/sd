package page

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
	Short: "List all page IDs",
	Run: func(cmd *cobra.Command, args []string) {
		// Connect to NATS.
		_, kv := natsconn.GetNATSConn()

		entries, err := kv.Keys()

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to list keys")
		}

		currentPageId := viper.GetString("current-page")

		// Extract unique instance IDs
		pageIDs := make(map[string]struct{}) // Use a map to ensure uniqueness

		for _, key := range entries {
			parts := strings.Split(key, ".")

			if len(parts) > 6 && parts[6] == "pages" && parts[7] != "current" { // TODO do this differently, it is a bit hacky.
				pageIDs[parts[7]] = struct{}{}
			}
		}

		// Create table using the helper function
		table := util.NewTable(cmd.OutOrStdout())
		table.SetHeader([]string{"Page ID", "CURRENT"})

		// Add rows
		for id := range pageIDs {
			currentMark := ""
			if id == currentPageId {
				currentMark = "CURRENT"
			}
			table.Append([]string{id, currentMark})
		}

		// Render the table
		table.Render()
	},
}

func init() {
	pageCmd.AddCommand(lsCmd)
}
