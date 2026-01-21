package cli

import "github.com/spf13/cobra"

var workOrdersCmd = &cobra.Command{
	Use:     "work-orders",
	Aliases: []string{"wo"},
	Short:   "View work orders",
	Long: `View work orders on the XBE platform.

Work orders group maintenance requirement sets for scheduling and completion.
They track the overall status, priority, and responsible party for maintenance work.

Commands:
  list    List work orders with filtering
  show    Show work order details`,
	Example: `  # List all work orders
  xbe view work-orders list

  # List work orders for my business units
  xbe view work-orders list --me

  # List work orders for a specific business unit
  xbe view work-orders list --bu-id 123

  # Filter by status
  xbe view work-orders list --status in_progress

  # Filter by priority
  xbe view work-orders list --priority urgent

  # Show work order details
  xbe view work-orders show 456`,
}

func init() {
	viewCmd.AddCommand(workOrdersCmd)
}
