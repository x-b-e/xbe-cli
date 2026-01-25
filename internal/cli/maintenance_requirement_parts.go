package cli

import "github.com/spf13/cobra"

var maintenanceRequirementPartsCmd = &cobra.Command{
	Use:     "maintenance-requirement-parts",
	Aliases: []string{"maintenance-requirement-part"},
	Short:   "View maintenance requirement parts",
	Long:    "Commands for viewing maintenance requirement parts.",
}

func init() {
	viewCmd.AddCommand(maintenanceRequirementPartsCmd)
}
