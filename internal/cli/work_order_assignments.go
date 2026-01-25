package cli

import "github.com/spf13/cobra"

var workOrderAssignmentsCmd = &cobra.Command{
	Use:     "work-order-assignments",
	Aliases: []string{"work-order-assignment"},
	Short:   "Browse work order assignments",
	Long: `Browse work order assignments.

Work order assignments link users to work orders.

Commands:
  list    List assignments with filtering and pagination
  show    Show full details of an assignment`,
	Example: `  # List assignments
  xbe view work-order-assignments list

  # Filter by work order
  xbe view work-order-assignments list --work-order 123

  # Filter by user
  xbe view work-order-assignments list --user 456

  # Show an assignment
  xbe view work-order-assignments show 789`,
}

func init() {
	viewCmd.AddCommand(workOrderAssignmentsCmd)
}
