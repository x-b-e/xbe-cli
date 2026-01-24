package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchInvoiceUnbatchingsCmd = &cobra.Command{
	Use:   "organization-invoices-batch-invoice-unbatchings",
	Short: "Unbatch organization invoices batch invoices",
	Long: `Unbatch organization invoices batch invoices.

Unbatchings transition organization invoices batch invoices from successful or failed
status to skipped.

Commands:
  create    Unbatch an organization invoices batch invoice`,
	Example: `  # Unbatch an organization invoices batch invoice
  xbe do organization-invoices-batch-invoice-unbatchings create \\
    --organization-invoices-batch-invoice 123 \\
    --comment "Unbatching for correction"`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchInvoiceUnbatchingsCmd)
}
