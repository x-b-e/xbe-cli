package cli

import "github.com/spf13/cobra"

var materialTransactionShiftAssignmentsCmd = &cobra.Command{
	Use:     "material-transaction-shift-assignments",
	Aliases: []string{"material-transaction-shift-assignment"},
	Short:   "Browse material transaction shift assignments",
	Long: `Browse material transaction shift assignments.

Shift assignments link material transactions to a tender job schedule shift
for a trucker and broker, optionally processing the assignments.

Commands:
  list    List material transaction shift assignments with filtering
  show    Show assignment details`,
	Example: `  # List assignments
  xbe view material-transaction-shift-assignments list

  # Filter by tender job schedule shift
  xbe view material-transaction-shift-assignments list --tender-job-schedule-shift 123

  # Show assignment details
  xbe view material-transaction-shift-assignments show 456`,
}

func init() {
	viewCmd.AddCommand(materialTransactionShiftAssignmentsCmd)
}
