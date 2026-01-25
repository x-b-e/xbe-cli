package cli

import "github.com/spf13/cobra"

var invoicesCmd = &cobra.Command{
	Use:   "invoices",
	Short: "Browse invoices",
	Long: `Browse invoices on the XBE platform.

Invoices capture billing details, status, and amounts for time card and
material transaction activity.

Note: Invoices are read-only in the API. Use invoice actions (invoice-addresses,
invoice-rejections, invoice-revisionables, invoice-revisionizings) to change
invoice status.

Commands:
  list    List invoices with filtering
  show    Show invoice details`,
	Example: `  # List invoices
  xbe view invoices list

  # Show an invoice
  xbe view invoices show 123`,
}

func init() {
	viewCmd.AddCommand(invoicesCmd)
}
