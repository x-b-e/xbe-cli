package cli

import "github.com/spf13/cobra"

var doCommentsCmd = &cobra.Command{
	Use:     "comments",
	Aliases: []string{"comment"},
	Short:   "Manage comments",
	Long:    "Commands for creating, updating, and deleting comments.",
}

func init() {
	doCmd.AddCommand(doCommentsCmd)
}
