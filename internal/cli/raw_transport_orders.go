package cli

import "github.com/spf13/cobra"

var rawTransportOrdersCmd = &cobra.Command{
	Use:     "raw-transport-orders",
	Aliases: []string{"raw-transport-order"},
	Short:   "Browse and view raw transport orders",
	Long: `Browse and view raw transport orders.

Raw transport orders capture imported TMW order data before it is normalized into
transport orders. Use them to audit raw import payloads, check import status,
and trace them to processed transport orders.

Commands:
  list    List raw transport orders with filtering
  show    View full details for a raw transport order`,
	Example: `  # List raw transport orders
  xbe view raw-transport-orders list

  # Filter by broker
  xbe view raw-transport-orders list --broker 123

  # Show a raw transport order
  xbe view raw-transport-orders show 456`,
}

func init() {
	viewCmd.AddCommand(rawTransportOrdersCmd)
}
