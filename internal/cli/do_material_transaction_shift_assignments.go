package cli

import "github.com/spf13/cobra"

var doMaterialTransactionShiftAssignmentsCmd = &cobra.Command{
	Use:     "material-transaction-shift-assignments",
	Aliases: []string{"material-transaction-shift-assignment"},
	Short:   "Manage material transaction shift assignments",
	Long: `Create material transaction shift assignments.

Assignments link material transactions to a tender job schedule shift.

Commands:
  create    Create a material transaction shift assignment`,
	Example: `  # Create a shift assignment
  xbe do material-transaction-shift-assignments create \
    --tender-job-schedule-shift 123 \
    --material-transaction-ids 456,789

  # Create with comment and validation override
  xbe do material-transaction-shift-assignments create \
    --tender-job-schedule-shift 123 \
    --material-transaction-ids 456,789 \
    --comment "Fix assignment" \
    --skip-material-transaction-shift-skew-validation`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionShiftAssignmentsCmd)
}
