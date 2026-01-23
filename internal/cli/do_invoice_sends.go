package cli

import "github.com/spf13/cobra"

var doInvoiceSendsCmd = &cobra.Command{
	Use:     "invoice-sends",
	Aliases: []string{"invoice-send"},
	Short:   "Send invoices",
	Long:    "Commands for sending invoices.",
}

func init() {
	doCmd.AddCommand(doInvoiceSendsCmd)
}
