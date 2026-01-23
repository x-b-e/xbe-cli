package cli

import "github.com/spf13/cobra"

var doCrewRequirementsCmd = &cobra.Command{
	Use:     "crew-requirements",
	Aliases: []string{"crew-requirement"},
	Short:   "Manage crew requirements",
	Long:    "Commands for creating, updating, and deleting crew requirements.",
}

func init() {
	doCmd.AddCommand(doCrewRequirementsCmd)
}
