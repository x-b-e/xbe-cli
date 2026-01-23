package cli

import "github.com/spf13/cobra"

var doMaterialTransactionPreloadsCmd = &cobra.Command{
	Use:     "material-transaction-preloads",
	Aliases: []string{"material-transaction-preload"},
	Short:   "Manage material transaction preloads",
	Long:    "Commands for creating and deleting material transaction preloads.",
}

func init() {
	doCmd.AddCommand(doMaterialTransactionPreloadsCmd)
}
