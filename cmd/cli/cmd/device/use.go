package device

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// useCmd represents the device use command
var useCmd = &cobra.Command{
	Use:   "use [serial]",
	Short: "Set the current scoped device",
	Args:  cobra.ExactArgs(1), // Require exactly one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		serial := args[0]

		// Validate the serial (optional, e.g., ensure it exists)
		// Add your validation logic here if necessary

		// Update the current device in the configuration
		viper.Set("current-device", serial)
		viper.Set("current-profile", "")
		viper.Set("current-page", "")
		viper.Set("current-button", "")
		viper.WriteConfig()

		fmt.Printf("Current scoped device set to: %s\n", serial)
		return nil
	},
}

func init() {
	deviceCmd.AddCommand(useCmd)
}
