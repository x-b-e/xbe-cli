package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchPdfGenerationsCmd = &cobra.Command{
	Use:   "organization-invoices-batch-pdf-generations",
	Short: "Manage organization invoices batch PDF generations",
	Long: `Manage organization invoices batch PDF generations.

These records track background PDF generation jobs for organization invoice
batches.

Commands:
  create    Create an organization invoices batch PDF generation`,
	Example: `  # Create a PDF generation
  xbe do organization-invoices-batch-pdf-generations create \
    --organization-invoices-batch 123 \
    --organization-pdf-template 456`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchPdfGenerationsCmd)
}
