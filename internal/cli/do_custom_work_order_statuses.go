package cli

import "github.com/spf13/cobra"

var doCustomWorkOrderStatusesCmd = &cobra.Command{
	Use:   "custom-work-order-statuses",
	Short: "Manage custom work order statuses",
	Long: `Create, update, and delete custom work order statuses.

Custom work order statuses allow brokers to define their own status labels
with colors that map to primary statuses (e.g., pending, in_progress, completed).

Commands:
  create    Create a new custom work order status
  update    Update an existing custom work order status
  delete    Delete a custom work order status`,
	Example: `  # Create a custom work order status
  xbe do custom-work-order-statuses create --label "Awaiting Parts" --primary-status pending --broker 123

  # Update a custom work order status
  xbe do custom-work-order-statuses update 456 --label "Parts Ordered" --color-hex "#FF5500"

  # Delete a custom work order status
  xbe do custom-work-order-statuses delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doCustomWorkOrderStatusesCmd)
}
