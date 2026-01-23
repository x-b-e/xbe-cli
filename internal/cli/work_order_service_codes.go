package cli

import "github.com/spf13/cobra"

var workOrderServiceCodesCmd = &cobra.Command{
	Use:   "work-order-service-codes",
	Short: "View work order service codes",
	Long: `View work order service codes on the XBE platform.

Work order service codes describe the service categories used on work orders
and are scoped to brokers.

Commands:
  list    List work order service codes
  show    Show work order service code details`,
	Example: `  # List work order service codes
  xbe view work-order-service-codes list

  # Filter by broker
  xbe view work-order-service-codes list --broker 123

  # Show a work order service code
  xbe view work-order-service-codes show 456`,
}

func init() {
	viewCmd.AddCommand(workOrderServiceCodesCmd)
}
