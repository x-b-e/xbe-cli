package cli

import "github.com/spf13/cobra"

var doMaintenanceRequirementPartsCmd = &cobra.Command{
	Use:     "maintenance-requirement-parts",
	Aliases: []string{"maintenance-requirement-part"},
	Short:   "Manage maintenance requirement parts",
	Long:    "Commands for creating, updating, and deleting maintenance requirement parts.",
}

func init() {
	doCmd.AddCommand(doMaintenanceRequirementPartsCmd)
}
