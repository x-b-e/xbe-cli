package cli

import "github.com/spf13/cobra"

var doInvoiceGenerationsCmd = &cobra.Command{
	Use:     "invoice-generations",
	Aliases: []string{"invoice-generation"},
	Short:   "Manage invoice generations",
	Long: `Create invoice generations.

Commands:
  create    Create an invoice generation`,
	Example: `  # Create an invoice generation
  xbe do invoice-generations create \
    --organization-type brokers \
    --organization-id 123 \
    --time-card-ids 456,789 \
    --note "End of week run"`,
}

func init() {
	doCmd.AddCommand(doInvoiceGenerationsCmd)
}
