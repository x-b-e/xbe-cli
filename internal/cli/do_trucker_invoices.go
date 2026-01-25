package cli

import "github.com/spf13/cobra"

var doTruckerInvoicesCmd = &cobra.Command{
	Use:     "trucker-invoices",
	Aliases: []string{"trucker-invoice"},
	Short:   "Manage trucker invoices",
	Long: `Manage trucker invoices on the XBE platform.

Commands:
  create    Create a trucker invoice
  update    Update a trucker invoice
  delete    Delete a trucker invoice`,
	Example: `  # Create a trucker invoice
  xbe do trucker-invoices create --buyer-type brokers --buyer 123 --seller-type truckers --seller 456 \\
    --invoice-date 2025-01-01 --due-on 2025-01-10 --adjustment-amount 0.00 --currency-code USD

  # Update notes
  xbe do trucker-invoices update 123 --notes "Updated notes"

  # Delete a trucker invoice
  xbe do trucker-invoices delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doTruckerInvoicesCmd)
}
