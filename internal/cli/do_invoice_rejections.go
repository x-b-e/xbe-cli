package cli

import "github.com/spf13/cobra"

var doInvoiceRejectionsCmd = &cobra.Command{
	Use:   "invoice-rejections",
	Short: "Reject sent invoices",
	Long: `Reject sent invoices on the XBE platform.

Rejecting an invoice transitions it from sent to rejected status.
Only sent invoices can be rejected.

Commands:
  create    Reject a sent invoice`,
	Example: `  # Reject a sent invoice
  xbe do invoice-rejections create --invoice 123 --comment "Missing documentation"`,
}

func init() {
	doCmd.AddCommand(doInvoiceRejectionsCmd)
}
