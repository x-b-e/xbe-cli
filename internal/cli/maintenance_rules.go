package cli

import "github.com/spf13/cobra"

var maintenanceRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "View maintenance requirement rules",
	Long: `View maintenance requirement rules.

Maintenance rules define the criteria for when maintenance should be performed
on equipment. Rules can be scoped to equipment classifications or business units
and specify maintenance types and schedules.

Commands:
  list    List rules with filtering
  show    View detailed rule information`,
	Example: `  # List all rules
  xbe view maintenance rules list

  # List only active rules
  xbe view maintenance rules list --active-only

  # Filter by business unit
  xbe view maintenance rules list --business-unit-id 123

  # View rule details
  xbe view maintenance rules show 456`,
}

func init() {
	maintenanceCmd.AddCommand(maintenanceRulesCmd)
}
