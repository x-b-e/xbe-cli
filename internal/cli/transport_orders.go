package cli

import "github.com/spf13/cobra"

var transportOrdersCmd = &cobra.Command{
	Use:   "transport-orders",
	Short: "View transport orders",
	Long: `View transport orders on the XBE platform.

This command provides a lightweight list of transport orders with
basic fields like customer, office, category, pickup/delivery, and miles.

Commands:
  list    List transport orders with filters`,
	Example: `  # List transport orders for a broker (defaults to today & tomorrow)
  xbe view transport-orders list --broker 297

  # Filter by order number (disables date filtering)
  xbe view transport-orders list --broker 297 --order-number 4114407

  # Filter by date window
  xbe view transport-orders list --broker 297 --start-on 2026-01-16 --end-on 2026-01-16

  # JSON output
  xbe view transport-orders list --broker 297 --json`,
}

func init() {
	viewCmd.AddCommand(transportOrdersCmd)
}
