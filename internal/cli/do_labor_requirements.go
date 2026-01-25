package cli

import "github.com/spf13/cobra"

var doLaborRequirementsCmd = &cobra.Command{
	Use:     "labor-requirements",
	Aliases: []string{"labor-requirement"},
	Short:   "Manage labor requirements",
	Long:    "Commands for creating, updating, and deleting labor requirements.",
}

func init() {
	doCmd.AddCommand(doLaborRequirementsCmd)
}
