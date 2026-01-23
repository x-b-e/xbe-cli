package cli

import "github.com/spf13/cobra"

var driverAssignmentRulesCmd = &cobra.Command{
	Use:     "driver-assignment-rules",
	Aliases: []string{"driver-assignment-rule"},
	Short:   "Browse driver assignment rules",
	Long: `Browse driver assignment rules.

Driver assignment rules define constraints or guidance used when assigning
drivers at specific levels (broker, project, shift, and more).

Commands:
  list    List driver assignment rules with filtering and pagination
  show    Show full details of a driver assignment rule`,
	Example: `  # List driver assignment rules
  xbe view driver-assignment-rules list

  # Filter by level
  xbe view driver-assignment-rules list --level-type Broker --level-id 123

  # Show a driver assignment rule
  xbe view driver-assignment-rules show 456`,
}

func init() {
	viewCmd.AddCommand(driverAssignmentRulesCmd)
}
