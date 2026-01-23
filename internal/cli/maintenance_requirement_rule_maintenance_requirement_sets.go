package cli

import "github.com/spf13/cobra"

var maintenanceRequirementRuleMaintenanceRequirementSetsCmd = &cobra.Command{
	Use:     "maintenance-requirement-rule-maintenance-requirement-sets",
	Aliases: []string{"maintenance-requirement-rule-maintenance-requirement-set"},
	Short:   "Browse maintenance requirement rule maintenance requirement sets",
	Long: `Browse maintenance requirement rule maintenance requirement sets on the XBE platform.

Maintenance requirement rule maintenance requirement sets link maintenance requirement
rules to template maintenance requirement sets.

Commands:
  list    List maintenance requirement rule maintenance requirement sets with filtering and pagination
  show    Show maintenance requirement rule maintenance requirement set details`,
	Example: `  # List maintenance requirement rule maintenance requirement sets
  xbe view maintenance-requirement-rule-maintenance-requirement-sets list

  # Show a maintenance requirement rule maintenance requirement set
  xbe view maintenance-requirement-rule-maintenance-requirement-sets show 123

  # Output as JSON
  xbe view maintenance-requirement-rule-maintenance-requirement-sets list --json`,
}

func init() {
	viewCmd.AddCommand(maintenanceRequirementRuleMaintenanceRequirementSetsCmd)
}
