package cli

import "github.com/spf13/cobra"

var doMaterialTransactionRejectionsCmd = &cobra.Command{
	Use:     "material-transaction-rejections",
	Aliases: []string{"material-transaction-rejection"},
	Short:   "Reject material transactions",
	Long:    "Commands for rejecting material transactions.",
}

func init() {
	doCmd.AddCommand(doMaterialTransactionRejectionsCmd)
}
