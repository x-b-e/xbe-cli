package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchInvoiceUnbatchingsCmd = &cobra.Command{
	Use:   "organization-invoices-batch-invoice-unbatchings",
	Short: "Browse organization invoices batch invoice unbatchings",
	Long: `Browse organization invoices batch invoice unbatchings on the XBE platform.

Unbatchings transition organization invoices batch invoices from successful or failed
status to skipped.

Commands:
  list    List organization invoices batch invoice unbatchings
  show    Show organization invoices batch invoice unbatching details`,
	Example: `  # List organization invoices batch invoice unbatchings
  xbe view organization-invoices-batch-invoice-unbatchings list

  # Show an organization invoices batch invoice unbatching
  xbe view organization-invoices-batch-invoice-unbatchings show 123`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchInvoiceUnbatchingsCmd)
}
