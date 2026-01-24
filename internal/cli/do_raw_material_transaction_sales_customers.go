package cli

import "github.com/spf13/cobra"

var doRawMaterialTransactionSalesCustomersCmd = &cobra.Command{
	Use:   "raw-material-transaction-sales-customers",
	Short: "Manage raw material transaction sales customers",
	Long: `Manage raw material transaction sales customers on the XBE platform.

Raw material transaction sales customers map raw sales customer identifiers to
customers and brokers for raw material transactions.

Commands:
  create    Create a raw material transaction sales customer
  update    Update a raw material transaction sales customer
  delete    Delete a raw material transaction sales customer`,
	Example: `  # Create a raw material transaction sales customer
  xbe do raw-material-transaction-sales-customers create --raw-sales-customer-id RAW-123 --customer 456

  # Update a raw material transaction sales customer
  xbe do raw-material-transaction-sales-customers update 789 --raw-sales-customer-id RAW-456

  # Delete a raw material transaction sales customer
  xbe do raw-material-transaction-sales-customers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doRawMaterialTransactionSalesCustomersCmd)
}
