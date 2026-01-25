package cli

import "github.com/spf13/cobra"

var timeCardInvoicesCmd = &cobra.Command{
	Use:   "time-card-invoices",
	Short: "Browse time card invoices",
	Long: `Browse time card invoices on the XBE platform.

Time card invoices link approved time cards to invoices for billing workflows.

Commands:
  list    List time card invoices
  show    Show time card invoice details`,
	Example: `  # List time card invoices
  xbe view time-card-invoices list

  # Show a time card invoice
  xbe view time-card-invoices show 123`,
}

func init() {
	viewCmd.AddCommand(timeCardInvoicesCmd)
}
