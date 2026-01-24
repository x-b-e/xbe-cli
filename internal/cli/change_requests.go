package cli

import "github.com/spf13/cobra"

var changeRequestsCmd = &cobra.Command{
	Use:     "change-requests",
	Aliases: []string{"change-request"},
	Short:   "View change requests",
	Long:    "Commands for viewing change requests.",
}

func init() {
	viewCmd.AddCommand(changeRequestsCmd)
}
