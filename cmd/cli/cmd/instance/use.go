package instance

import (
	"fmt"
	"sd/cmd/cli/pkg/config"

	"github.com/spf13/cobra"
)

// useCmd represents the instance use command
var useCmd = &cobra.Command{
	Use:   "use [uuid]",
	Short: "Set the current scoped instance",
	Args:  cobra.ExactArgs(1), // Require exactly one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		uuid := args[0]

		// Load the current configuration
		conf, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %v", err)
		}

		// Validate the UUID (optional, e.g., ensure it exists)
		// Add your validation logic here if necessary

		// Update the current instance in the configuration
		conf.CurrentInstance = uuid
		if err := config.SaveConfig(conf); err != nil {
			return fmt.Errorf("failed to save configuration: %v", err)
		}

		fmt.Printf("Current scoped instance set to: %s\n", uuid)
		return nil
	},
}

func init() {
	instanceCmd.AddCommand(useCmd)
}
