package cli

import "github.com/spf13/cobra"

var doProjectPhasesCmd = &cobra.Command{
	Use:     "project-phases",
	Aliases: []string{"project-phase"},
	Short:   "Manage project phases",
	Long:    "Commands for creating, updating, and deleting project phases.",
}

func init() {
	doCmd.AddCommand(doProjectPhasesCmd)
}
