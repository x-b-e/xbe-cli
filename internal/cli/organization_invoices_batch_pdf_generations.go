package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchPdfGenerationsCmd = &cobra.Command{
	Use:   "organization-invoices-batch-pdf-generations",
	Short: "Browse organization invoices batch PDF generations",
	Long: `Browse organization invoices batch PDF generations on the XBE platform.

These records track background PDF generation jobs for organization invoice
batches, including status and generated files.

Commands:
  list          List organization invoices batch PDF generations with filtering
  show          Show organization invoices batch PDF generation details
  download-all  Download all completed PDFs as a ZIP archive`,
	Example: `  # List PDF generations
  xbe view organization-invoices-batch-pdf-generations list

  # Show a PDF generation
  xbe view organization-invoices-batch-pdf-generations show 123

  # Download all PDFs for a generation
  xbe view organization-invoices-batch-pdf-generations download-all 123 --output ./batch_pdfs.zip`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchPdfGenerationsCmd)
}
