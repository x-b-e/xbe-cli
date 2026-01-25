package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchPdfFilesCmd = &cobra.Command{
	Use:     "organization-invoices-batch-pdf-files",
	Aliases: []string{"organization-invoices-batch-pdf-file"},
	Short:   "Browse organization invoices batch PDF files",
	Long: `Browse organization invoices batch PDF files.

Organization invoices batch PDF files represent generated invoice PDFs produced
as part of an organization invoice batch PDF generation.

Commands:
  list    List organization invoices batch PDF files
  show    Show organization invoices batch PDF file details`,
	Example: `  # List PDF files for a batch
  xbe view organization-invoices-batch-pdf-files list --pdf-generation 123

  # Show a PDF file
  xbe view organization-invoices-batch-pdf-files show 456`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchPdfFilesCmd)
}
