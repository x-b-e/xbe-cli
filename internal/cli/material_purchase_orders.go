package cli

import "github.com/spf13/cobra"

var materialPurchaseOrdersCmd = &cobra.Command{
	Use:   "material-purchase-orders",
	Short: "Browse and view material purchase orders",
	Long: `Browse and view material purchase orders on the XBE platform.

Material purchase orders define planned material quantities and constraints for
brokers, suppliers, and customers. They can be used to issue releases and track
fulfillment against ordered quantities.

Statuses:
  editing   Draft order, editable
  approved  Approved for release creation
  closed    Closed or complete

Commands:
  list    List material purchase orders with filtering
  show    View full details for a material purchase order`,
	Example: `  # List recent material purchase orders
  xbe view material-purchase-orders list

  # Filter by supplier and status
  xbe view material-purchase-orders list --material-supplier 123 --status approved

  # View a specific purchase order
  xbe view material-purchase-orders show 456`,
}

func init() {
	viewCmd.AddCommand(materialPurchaseOrdersCmd)
}
