package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchInvoicesCmd = &cobra.Command{
	Use:     "organization-invoices-batch-invoices",
	Aliases: []string{"organization-invoices-batch-invoice"},
	Short:   "View organization invoices batch invoices",
	Long: `View organization invoices batch invoices on the XBE platform.

Organization invoices batch invoices track invoices included in organization
invoice batches and their batch processing status.

Commands:
  list    List organization invoices batch invoices
  show    Show organization invoices batch invoice details`,
	Example: `  # List batch invoices
  xbe view organization-invoices-batch-invoices list

  # Filter by batch
  xbe view organization-invoices-batch-invoices list --organization-invoices-batch 123

  # Filter by invoice ID
  xbe view organization-invoices-batch-invoices list --invoice-id 456

  # Show a batch invoice
  xbe view organization-invoices-batch-invoices show 789`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchInvoicesCmd)
}
