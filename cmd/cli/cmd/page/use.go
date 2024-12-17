package page

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// useCmd represents the page use command
var useCmd = &cobra.Command{
	Use:   "use [serial]",
	Short: "Set the current scoped page",
	Args:  cobra.ExactArgs(1), // Require exactly one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		serial := args[0]

		// Validate the serial (optional, e.g., ensure it exists)
		// Add your validation logic here if necessary

		// Update the current page in the configuration
		viper.Set("current-page", serial)
		viper.Set("current-button", "")
		viper.WriteConfig()

		fmt.Printf("Current scoped page set to: %s\n", serial)
		return nil
	},
}

func init() {
	pageCmd.AddCommand(useCmd)
}
