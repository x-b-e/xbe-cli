package cli

import "github.com/spf13/cobra"

var doProffersCmd = &cobra.Command{
	Use:     "proffers",
	Aliases: []string{"proffer"},
	Short:   "Manage proffers",
	Long:    "Commands for creating, updating, and deleting proffers.",
}

func init() {
	doCmd.AddCommand(doProffersCmd)
}
