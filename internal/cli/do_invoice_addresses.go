package cli

import "github.com/spf13/cobra"

var doInvoiceAddressesCmd = &cobra.Command{
	Use:   "invoice-addresses",
	Short: "Address rejected invoices",
	Long: `Address rejected invoices on the XBE platform.

Addressing an invoice transitions it from rejected to addressed status.
Only rejected invoices can be addressed.

Commands:
  create    Address a rejected invoice`,
	Example: `  # Address a rejected invoice
  xbe do invoice-addresses create --invoice 123 --comment "Resolved dispute"`,
}

func init() {
	doCmd.AddCommand(doInvoiceAddressesCmd)
}
