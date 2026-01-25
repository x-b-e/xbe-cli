package cli

import "github.com/spf13/cobra"

var doMaintenanceRequirementRuleMaintenanceRequirementSetsCmd = &cobra.Command{
	Use:     "maintenance-requirement-rule-maintenance-requirement-sets",
	Aliases: []string{"maintenance-requirement-rule-maintenance-requirement-set"},
	Short:   "Manage maintenance requirement rule maintenance requirement sets",
	Long: `Manage maintenance requirement rule maintenance requirement sets.

Maintenance requirement rule maintenance requirement sets link maintenance requirement
rules to template maintenance requirement sets.

Commands:
  create    Create a maintenance requirement rule maintenance requirement set
  delete    Delete a maintenance requirement rule maintenance requirement set`,
	Example: `  # Create a maintenance requirement rule maintenance requirement set
  xbe do maintenance-requirement-rule-maintenance-requirement-sets create --maintenance-requirement-rule 123 --maintenance-requirement-set 456

  # Delete a maintenance requirement rule maintenance requirement set
  xbe do maintenance-requirement-rule-maintenance-requirement-sets delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doMaintenanceRequirementRuleMaintenanceRequirementSetsCmd)
}
