package profile

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// useCmd represents the profile use command
var useCmd = &cobra.Command{
	Use:   "use [serial]",
	Short: "Set the current scoped profile",
	Args:  cobra.ExactArgs(1), // Require exactly one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		serial := args[0]

		// Validate the serial (optional, e.g., ensure it exists)
		// Add your validation logic here if necessary

		// Update the current profile in the configuration
		viper.Set("current-profile", serial)
		viper.Set("current-page", "")
		viper.Set("current-button", "")
		viper.WriteConfig()

		fmt.Printf("Current scoped profile set to: %s\n", serial)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(useCmd)
}
