package cli

import "github.com/spf13/cobra"

var doMaterialTransactionsCmd = &cobra.Command{
	Use:     "material-transactions",
	Aliases: []string{"material-transaction", "mtxn", "mtxns"},
	Short:   "Manage material transactions",
	Long:    `Create and delete material transactions.`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionsCmd)
}
