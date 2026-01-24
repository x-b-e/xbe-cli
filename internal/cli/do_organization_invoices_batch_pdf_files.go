package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchPdfFilesCmd = &cobra.Command{
	Use:     "organization-invoices-batch-pdf-files",
	Aliases: []string{"organization-invoices-batch-pdf-file"},
	Short:   "Download organization invoices batch PDF files",
	Long: `Download organization invoices batch PDF files.

Organization invoices batch PDF files are generated invoice PDFs from a batch.

Commands:
  download    Download an organization invoices batch PDF file`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchPdfFilesCmd)
}
