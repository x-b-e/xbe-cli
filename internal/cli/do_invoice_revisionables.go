package cli

import "github.com/spf13/cobra"

var doInvoiceRevisionablesCmd = &cobra.Command{
	Use:   "invoice-revisionables",
	Short: "Mark invoices as revisionable",
	Long: `Mark invoices as revisionable on the XBE platform.

This action transitions an invoice to revisionable status.
Only exported, non-exportable, or revised invoices can be marked revisionable.

Commands:
  create    Mark an invoice as revisionable`,
	Example: `  # Mark an invoice as revisionable
  xbe do invoice-revisionables create --invoice 123 --comment "Needs updates"`,
}

func init() {
	doCmd.AddCommand(doInvoiceRevisionablesCmd)
}
