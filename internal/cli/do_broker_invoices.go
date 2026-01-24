package cli

import "github.com/spf13/cobra"

var doBrokerInvoicesCmd = &cobra.Command{
	Use:     "broker-invoices",
	Aliases: []string{"broker-invoice"},
	Short:   "Manage broker invoices",
	Long: `Create, update, and delete broker invoices.

Commands:
  create    Create a broker invoice
  update    Update a broker invoice
  delete    Delete a broker invoice`,
	Example: `  # Create a broker invoice
  xbe do broker-invoices create --customer 123 --broker 456 --invoice-date 2025-01-01 --due-on 2025-01-31 --adjustment-amount 0 --currency-code USD

  # Update notes
  xbe do broker-invoices update 123 --notes "Updated"

  # Delete a broker invoice
  xbe do broker-invoices delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerInvoicesCmd)
}
