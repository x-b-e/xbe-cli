package cli

import "github.com/spf13/cobra"

var doMaintenanceRequirementSetMaintenanceRequirementsCmd = &cobra.Command{
	Use:     "maintenance-requirement-set-maintenance-requirements",
	Aliases: []string{"maintenance-requirement-set-maintenance-requirement"},
	Short:   "Manage maintenance requirement set maintenance requirements",
	Long: `Create, update, and delete maintenance requirement set maintenance requirements.

These records link maintenance requirements to maintenance requirement sets.

Commands:
  create    Create a record
  update    Update a record
  delete    Delete a record`,
}

func init() {
	doCmd.AddCommand(doMaintenanceRequirementSetMaintenanceRequirementsCmd)
}
