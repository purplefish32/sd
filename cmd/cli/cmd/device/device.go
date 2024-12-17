package device

import (
	"github.com/spf13/cobra"
)

// deviceCmd represents the device command
var deviceCmd = &cobra.Command{}

func NewCmd() *cobra.Command {
	// Create a new command group for 'instance' and add 'ls' as a subcommand
	cmd := &cobra.Command{
		Use:   "device",
		Short: "Commands related to devices",
	}

	cmd.AddCommand(lsCmd)
	cmd.AddCommand(useCmd)

	return cmd
}
