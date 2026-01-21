package cli

import "github.com/spf13/cobra"

var certificationRequirementsCmd = &cobra.Command{
	Use:     "certification-requirements",
	Aliases: []string{"certification-requirement"},
	Short:   "View certification requirements",
	Long:    "Commands for viewing certification requirements.",
}

func init() {
	viewCmd.AddCommand(certificationRequirementsCmd)
}
