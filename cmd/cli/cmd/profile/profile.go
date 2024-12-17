package profile

import (
	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{}

func NewCmd() *cobra.Command {
	// Create a new command group for 'instance' and add 'ls' as a subcommand
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Commands related to profiles",
	}

	cmd.AddCommand(lsCmd)
	cmd.AddCommand(useCmd)

	return cmd
}
