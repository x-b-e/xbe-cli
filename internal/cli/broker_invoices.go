package cli

import "github.com/spf13/cobra"

var brokerInvoicesCmd = &cobra.Command{
	Use:     "broker-invoices",
	Aliases: []string{"broker-invoice"},
	Short:   "Browse broker invoices",
	Long: `Browse broker invoices.

Broker invoices (customer invoices) represent customer-facing invoices
issued by brokers. Use these commands to list and inspect invoice details.

Commands:
  list    List broker invoices with filtering
  show    Show broker invoice details`,
	Example: `  # List broker invoices
  xbe view broker-invoices list

  # Show a broker invoice
  xbe view broker-invoices show 123`,
}

func init() {
	viewCmd.AddCommand(brokerInvoicesCmd)
}
