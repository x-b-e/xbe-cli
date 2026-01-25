package cli

import "github.com/spf13/cobra"

var rawMaterialTransactionsCmd = &cobra.Command{
	Use:     "raw-material-transactions",
	Aliases: []string{"raw-material-transaction"},
	Short:   "Browse and view raw material transactions",
	Long: `Browse and view raw material transactions.

Raw material transactions capture the inbound ticket records from material sites
before they are normalized into material transactions. Use them to audit raw data
and trace how a ticket maps to a processed transaction.

Commands:
  list    List raw material transactions with filtering
  show    View full details for a raw material transaction`,
	Example: `  # List recent raw material transactions
  xbe view raw-material-transactions list

  # Filter by transaction date
  xbe view raw-material-transactions list --date 2025-01-15

  # Filter by ticket number
  xbe view raw-material-transactions list --ticket-number T12345

  # View a specific raw material transaction
  xbe view raw-material-transactions show 123`,
}

func init() {
	viewCmd.AddCommand(rawMaterialTransactionsCmd)
}
