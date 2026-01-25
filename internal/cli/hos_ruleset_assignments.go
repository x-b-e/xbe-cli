package cli

import "github.com/spf13/cobra"

var hosRulesetAssignmentsCmd = &cobra.Command{
	Use:     "hos-ruleset-assignments",
	Aliases: []string{"hos-ruleset-assignment"},
	Short:   "Browse HOS ruleset assignments",
	Long: `Browse HOS ruleset assignments.

HOS ruleset assignments track which HOS rule set applies to a driver at a
specific time, including mid-day rule set changes.

Commands:
  list    List HOS ruleset assignments with filtering and pagination
  show    Show full details of a HOS ruleset assignment`,
	Example: `  # List HOS ruleset assignments
  xbe view hos-ruleset-assignments list

  # Filter by driver
  xbe view hos-ruleset-assignments list --driver 123

  # Show a HOS ruleset assignment
  xbe view hos-ruleset-assignments show 456`,
}

func init() {
	viewCmd.AddCommand(hosRulesetAssignmentsCmd)
}
