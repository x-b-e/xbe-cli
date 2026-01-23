package cli

import "github.com/spf13/cobra"

var invoiceGenerationsCmd = &cobra.Command{
	Use:     "invoice-generations",
	Aliases: []string{"invoice-generation"},
	Short:   "Browse invoice generations",
	Long: `Browse invoice generations.

Invoice generations track batch processing runs that create invoices for
an organization over selected time cards.

Commands:
  list    List invoice generations with filtering
  show    Show invoice generation details`,
	Example: `  # List invoice generations
  xbe view invoice-generations list

  # Show an invoice generation
  xbe view invoice-generations show 123`,
}

func init() {
	viewCmd.AddCommand(invoiceGenerationsCmd)
}
