package cli

import "github.com/spf13/cobra"

var doRawMaterialTransactionsCmd = &cobra.Command{
	Use:     "raw-material-transactions",
	Aliases: []string{"raw-material-transaction"},
	Short:   "Manage raw material transactions",
	Long: `Update raw material transactions.

Note: Only admin users can update raw material transactions.`,
}

func init() {
	doCmd.AddCommand(doRawMaterialTransactionsCmd)
}
