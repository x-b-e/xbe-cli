package cli

import "github.com/spf13/cobra"

var doChangeRequestsCmd = &cobra.Command{
	Use:     "change-requests",
	Aliases: []string{"change-request"},
	Short:   "Manage change requests",
	Long:    "Commands for creating, updating, and deleting change requests.",
}

func init() {
	doCmd.AddCommand(doChangeRequestsCmd)
}
