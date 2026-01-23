package cli

import "github.com/spf13/cobra"

var doMaintenanceRequirementMaintenanceRequirementPartsCmd = &cobra.Command{
	Use:     "maintenance-requirement-maintenance-requirement-parts",
	Aliases: []string{"maintenance-requirement-maintenance-requirement-part"},
	Short:   "Manage maintenance requirement parts",
	Long: `Create, update, and delete maintenance requirement part links.

Commands:
  create    Create a maintenance requirement part link
  update    Update a maintenance requirement part link
  delete    Delete a maintenance requirement part link`,
}

func init() {
	doCmd.AddCommand(doMaintenanceRequirementMaintenanceRequirementPartsCmd)
}
