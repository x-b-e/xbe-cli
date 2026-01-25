package cli

import "github.com/spf13/cobra"

var truckerInvoicesCmd = &cobra.Command{
	Use:     "trucker-invoices",
	Aliases: []string{"trucker-invoice"},
	Short:   "Browse trucker invoices",
	Long: `Browse trucker invoices on the XBE platform.

Trucker invoices represent bills issued to brokers for trucker time cards.

Commands:
  list    List trucker invoices with filtering
  show    Show trucker invoice details`,
	Example: `  # List trucker invoices
  xbe view trucker-invoices list

  # Show a trucker invoice
  xbe view trucker-invoices show 123`,
}

func init() {
	viewCmd.AddCommand(truckerInvoicesCmd)
}
