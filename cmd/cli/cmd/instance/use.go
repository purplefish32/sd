package instance

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// useCmd represents the instance use command
var useCmd = &cobra.Command{
	Use:   "use [uuid]",
	Short: "Set the current scoped instance",
	Args:  cobra.ExactArgs(1), // Require exactly one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		uuid := args[0]

		// Validate the UUID (optional, e.g., ensure it exists)
		// Add your validation logic here if necessary

		// Update the current instance in the configuration
		viper.Set("current-instance", uuid)
		viper.Set("current-device", "")
		viper.Set("current-profile", "")
		viper.Set("current-page", "")
		viper.Set("current-button", "")

		err := viper.WriteConfig()

		if err != nil {
			fmt.Printf("Error: %+v", err)

		}

		fmt.Printf("Current scoped instance set to: %s\n", uuid)
		return nil
	},
}

func init() {
	instanceCmd.AddCommand(useCmd)
}
