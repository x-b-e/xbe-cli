package cli

import "github.com/spf13/cobra"

var doInvoiceExportsCmd = &cobra.Command{
	Use:     "invoice-exports",
	Aliases: []string{"invoice-export"},
	Short:   "Export invoices",
	Long: `Export invoices.

Invoice exports move approved or batched invoices to exported status and
create a revision.

Commands:
  create    Export an invoice`,
	Example: `  # Export an invoice
  xbe do invoice-exports create --invoice 123 --comment "Sent to accounting"`,
}

func init() {
	doCmd.AddCommand(doInvoiceExportsCmd)
}
