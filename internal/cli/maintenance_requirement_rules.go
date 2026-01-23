package cli

import "github.com/spf13/cobra"

var maintenanceRequirementRulesCmd = &cobra.Command{
	Use:     "maintenance-requirement-rules",
	Aliases: []string{"maintenance-requirement-rule"},
	Short:   "Browse maintenance requirement rules",
	Long: `Browse maintenance requirement rules.

Maintenance requirement rules define maintenance or inspection requirements for
specific equipment, equipment classifications, or business units.

Commands:
  list    List maintenance requirement rules with filtering and pagination
  show    Show full details of a maintenance requirement rule`,
	Example: `  # List maintenance requirement rules
  xbe view maintenance-requirement-rules list

  # Filter by equipment classification
  xbe view maintenance-requirement-rules list --equipment-classification 123

  # Show a maintenance requirement rule
  xbe view maintenance-requirement-rules show 456`,
}

func init() {
	viewCmd.AddCommand(maintenanceRequirementRulesCmd)
}
