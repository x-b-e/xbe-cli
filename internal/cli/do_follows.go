package cli

import "github.com/spf13/cobra"

var doFollowsCmd = &cobra.Command{
	Use:     "follows",
	Aliases: []string{"follow"},
	Short:   "Manage follows",
	Long:    "Commands for creating, updating, and deleting follows.",
}

func init() {
	doCmd.AddCommand(doFollowsCmd)
}
