package cli

import "github.com/spf13/cobra"

var rawMaterialTransactionSalesCustomersCmd = &cobra.Command{
	Use:     "raw-material-transaction-sales-customers",
	Aliases: []string{"raw-material-transaction-sales-customer"},
	Short:   "Browse raw material transaction sales customers",
	Long: `Browse raw material transaction sales customers on the XBE platform.

Raw material transaction sales customers map raw sales customer identifiers to
customers and brokers for raw material transactions.

Commands:
  list    List raw material transaction sales customers with filtering
  show    Show raw material transaction sales customer details`,
	Example: `  # List raw material transaction sales customers
  xbe view raw-material-transaction-sales-customers list

  # Show a raw material transaction sales customer
  xbe view raw-material-transaction-sales-customers show 123`,
}

func init() {
	viewCmd.AddCommand(rawMaterialTransactionSalesCustomersCmd)
}
