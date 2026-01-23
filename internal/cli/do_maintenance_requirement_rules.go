package cli

import "github.com/spf13/cobra"

var doMaintenanceRequirementRulesCmd = &cobra.Command{
	Use:     "maintenance-requirement-rules",
	Aliases: []string{"maintenance-requirement-rule"},
	Short:   "Manage maintenance requirement rules",
	Long: `Create, update, and delete maintenance requirement rules.

Maintenance requirement rules define maintenance or inspection requirements for
specific equipment, equipment classifications, or business units.`,
}

func init() {
	doCmd.AddCommand(doMaintenanceRequirementRulesCmd)
}
