package cli

import "github.com/spf13/cobra"

var doMaterialTransactionSubmissionsCmd = &cobra.Command{
	Use:     "material-transaction-submissions",
	Aliases: []string{"material-transaction-submission"},
	Short:   "Submit material transactions",
	Long:    "Commands for submitting material transactions.",
}

func init() {
	doCmd.AddCommand(doMaterialTransactionSubmissionsCmd)
}
