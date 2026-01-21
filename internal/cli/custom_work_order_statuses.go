package cli

import "github.com/spf13/cobra"

var customWorkOrderStatusesCmd = &cobra.Command{
	Use:   "custom-work-order-statuses",
	Short: "View custom work order statuses",
	Long: `View custom work order statuses on the XBE platform.

Custom work order statuses allow brokers to define their own status labels
with colors that map to primary statuses (e.g., pending, in_progress, completed).

Commands:
  list    List custom work order statuses`,
	Example: `  # List custom work order statuses
  xbe view custom-work-order-statuses list

  # Filter by broker
  xbe view custom-work-order-statuses list --broker 123

  # Filter by primary status
  xbe view custom-work-order-statuses list --primary-status pending

  # Output as JSON
  xbe view custom-work-order-statuses list --json`,
}

func init() {
	viewCmd.AddCommand(customWorkOrderStatusesCmd)
}
