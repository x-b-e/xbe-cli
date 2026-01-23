package cli

import "github.com/spf13/cobra"

var laborRequirementsCmd = &cobra.Command{
	Use:     "labor-requirements",
	Aliases: []string{"labor-requirement"},
	Short:   "View labor requirements",
	Long:    "Commands for viewing labor requirements.",
}

func init() {
	viewCmd.AddCommand(laborRequirementsCmd)
}
