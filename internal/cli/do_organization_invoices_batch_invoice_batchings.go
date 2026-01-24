package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchInvoiceBatchingsCmd = &cobra.Command{
	Use:     "organization-invoices-batch-invoice-batchings",
	Aliases: []string{"organization-invoices-batch-invoice-batching"},
	Short:   "Batch organization invoices batch invoices",
	Long: `Batch organization invoices batch invoices.

Organization invoices batch invoice batchings mark skipped or failed batch invoices as successful.

Commands:
  create    Batch an organization invoices batch invoice`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchInvoiceBatchingsCmd)
}
