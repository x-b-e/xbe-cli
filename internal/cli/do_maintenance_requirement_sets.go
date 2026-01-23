package cli

import "github.com/spf13/cobra"

var doMaintenanceRequirementSetsCmd = &cobra.Command{
	Use:     "maintenance-requirement-sets",
	Aliases: []string{"maintenance-requirement-set"},
	Short:   "Manage maintenance requirement sets",
	Long: `Create, update, and delete maintenance requirement sets.

Commands:
  create    Create a maintenance requirement set
  update    Update a maintenance requirement set
  delete    Delete a maintenance requirement set`,
}

func init() {
	doCmd.AddCommand(doMaintenanceRequirementSetsCmd)
}
