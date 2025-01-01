package button

import (
	"fmt"
	"sd/pkg/natsconn"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// useCmd represents the profile use command
var getCmd = &cobra.Command{
	Use:   "get [buttonId]",
	Short: "Get the button",
	Args:  cobra.ExactArgs(1), // Require exactly one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		buttonID := args[0]

		currentInstance := viper.GetString("current-instance")
		currentDevice := viper.GetString("current-device")
		currentProfile := viper.GetString("current-profile")
		currentPage := viper.GetString("current-page")

		// Connect to NATS.
		_, kv := natsconn.GetNATSConn()

		entry, _ := kv.Get("instances." + currentInstance + ".devices." + currentDevice + ".profiles." + currentProfile + ".pages." + currentPage + ".buttons." + buttonID)

		// Validate the serial (optional, e.g., ensure it exists)
		// Add your validation logic here if necessary

		// Update the current profile in the configuration

		fmt.Print(string(entry.Value()))

		return nil
	},
}

func init() {
	buttonCmd.AddCommand(getCmd)
}
