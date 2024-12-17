package instance

import (
	"github.com/spf13/cobra"
)

// instanceCmd represents the instance command
var instanceCmd = &cobra.Command{}

func NewCmd() *cobra.Command {
	// Create a new command group for 'instance' and add 'ls' as a subcommand
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "Commands related to instances",
	}

	cmd.AddCommand(lsCmd)
	cmd.AddCommand(useCmd)

	return cmd
}
