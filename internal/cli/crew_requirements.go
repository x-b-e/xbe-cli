package cli

import "github.com/spf13/cobra"

var crewRequirementsCmd = &cobra.Command{
	Use:     "crew-requirements",
	Aliases: []string{"crew-requirement"},
	Short:   "View crew requirements",
	Long:    "Commands for viewing crew requirements.",
}

func init() {
	viewCmd.AddCommand(crewRequirementsCmd)
}
