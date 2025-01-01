package profile

import (
	"fmt"
	"sd/pkg/profiles"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// useCmd represents the profile use command
var useCmd = &cobra.Command{
	Use:   "use [profileID]",
	Short: "Set the current scoped profile",
	Args:  cobra.ExactArgs(1), // Require exactly one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		profileID := args[0]

		// Validate the profileID (optional, e.g., ensure it exists)
		// Add your validation logic here if necessary

		// Update the current profile in the configuration
		viper.Set("current-profile", profileID)
		viper.Set("current-page", "")
		viper.Set("current-button", "")
		viper.WriteConfig()

		currentInstanceID := viper.GetString("current-instance")
		currentDeviceID := viper.GetString("current-device")

		profiles.SetCurrentProfile(currentInstanceID, currentDeviceID, profileID)

		fmt.Printf("Current scoped profile set to: %s\n", profileID)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(useCmd)
}
