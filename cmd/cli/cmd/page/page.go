package page

import (
	"github.com/spf13/cobra"
)

// pageCmd represents the page command
var pageCmd = &cobra.Command{}

func NewCmd() *cobra.Command {
	// Create a new command group for 'instance' and add 'ls' as a subcommand
	cmd := &cobra.Command{
		Use:   "page",
		Short: "Manage pages",
	}

	cmd.AddCommand(lsCmd)
	cmd.AddCommand(useCmd)

	return cmd
}
