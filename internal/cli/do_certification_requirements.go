package cli

import "github.com/spf13/cobra"

var doCertificationRequirementsCmd = &cobra.Command{
	Use:     "certification-requirements",
	Aliases: []string{"certification-requirement"},
	Short:   "Manage certification requirements",
	Long:    "Commands for creating, updating, and deleting certification requirements.",
}

func init() {
	doCmd.AddCommand(doCertificationRequirementsCmd)
}
