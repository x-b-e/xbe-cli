package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchInvoiceBatchingsCmd = &cobra.Command{
	Use:     "organization-invoices-batch-invoice-batchings",
	Aliases: []string{"organization-invoices-batch-invoice-batching"},
	Short:   "View organization invoices batch invoice batchings",
	Long: `View organization invoices batch invoice batchings.

Organization invoices batch invoice batchings mark skipped or failed batch invoices as successful.

Commands:
  list    List organization invoices batch invoice batchings
  show    Show organization invoices batch invoice batching details`,
	Example: `  # List organization invoices batch invoice batchings
  xbe view organization-invoices-batch-invoice-batchings list

  # Show an organization invoices batch invoice batching
  xbe view organization-invoices-batch-invoice-batchings show 123

  # Output JSON
  xbe view organization-invoices-batch-invoice-batchings list --json`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchInvoiceBatchingsCmd)
}
