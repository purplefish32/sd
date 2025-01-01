package button

import (
	"github.com/spf13/cobra"
)

// instanceCmd represents the instance command
var buttonCmd = &cobra.Command{}

func NewCmd() *cobra.Command {
	// Create a new command group for 'instance' and add 'ls' as a subcommand
	cmd := &cobra.Command{
		Use:   "button",
		Short: "Commands related to buttons",
	}

	cmd.AddCommand(getCmd)

	return cmd
}
