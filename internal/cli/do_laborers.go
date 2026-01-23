package cli

import "github.com/spf13/cobra"

var doLaborersCmd = &cobra.Command{
	Use:     "laborers",
	Aliases: []string{"laborer"},
	Short:   "Manage laborers",
	Long:    "Commands for creating, updating, and deleting laborers.",
}

func init() {
	doCmd.AddCommand(doLaborersCmd)
}
