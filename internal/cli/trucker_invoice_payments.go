package cli

import "github.com/spf13/cobra"

var truckerInvoicePaymentsCmd = &cobra.Command{
	Use:     "trucker-invoice-payments",
	Aliases: []string{"trucker-invoice-payment"},
	Short:   "View trucker invoice payments",
	Long: `View trucker invoice payments.

Trucker invoice payments are QuickBooks bill payments linked to truckers.

Commands:
  list    List trucker invoice payments
  show    Show trucker invoice payment details`,
}

func init() {
	viewCmd.AddCommand(truckerInvoicePaymentsCmd)
}
