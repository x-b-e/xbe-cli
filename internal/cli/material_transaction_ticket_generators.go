package cli

import "github.com/spf13/cobra"

var materialTransactionTicketGeneratorsCmd = &cobra.Command{
	Use:     "material-transaction-ticket-generators",
	Aliases: []string{"material-transaction-ticket-generator"},
	Short:   "View material transaction ticket generators",
	Long: `View material transaction ticket generators.

Material transaction ticket generators define ticket numbering format rules for
brokers or material suppliers.

Commands:
  list    List material transaction ticket generators
  show    Show material transaction ticket generator details`,
	Example: `  # List ticket generators
  xbe view material-transaction-ticket-generators list

  # Show a ticket generator
  xbe view material-transaction-ticket-generators show 123`,
}

func init() {
	viewCmd.AddCommand(materialTransactionTicketGeneratorsCmd)
}
