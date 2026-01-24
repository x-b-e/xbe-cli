package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchInvoiceStatusChangesCmd = &cobra.Command{
	Use:     "organization-invoices-batch-invoice-status-changes",
	Aliases: []string{"organization-invoices-batch-invoice-status-change"},
	Short:   "View organization invoices batch invoice status changes",
	Long: `View organization invoices batch invoice status changes on the XBE platform.

Organization invoices batch invoice status changes track status transitions for
invoices within organization invoice batches.

Commands:
  list    List organization invoices batch invoice status changes
  show    Show organization invoices batch invoice status change details`,
	Example: `  # List status changes
  xbe view organization-invoices-batch-invoice-status-changes list

  # Filter by batch invoice
  xbe view organization-invoices-batch-invoice-status-changes list --organization-invoices-batch-invoice 123

  # Show a status change
  xbe view organization-invoices-batch-invoice-status-changes show 456`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchInvoiceStatusChangesCmd)
}
