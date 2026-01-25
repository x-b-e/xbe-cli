package cli

import "github.com/spf13/cobra"

var doInvoiceRevisionizingsCmd = &cobra.Command{
	Use:   "invoice-revisionizings",
	Short: "Revise invoices",
	Long: `Revise invoices on the XBE platform.

This action transitions invoice status to revised.
Only revisionable, exported, or non-exportable invoices can be revised.
Bulk revisionizing is required by the API.

Commands:
  create    Revise an invoice`,
	Example: `  # Revise an invoice (bulk required)
  xbe do invoice-revisionizings create --invoice 123 --comment "Bulk revision" --in-bulk`,
}

func init() {
	doCmd.AddCommand(doInvoiceRevisionizingsCmd)
}
