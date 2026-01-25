package cli

import "github.com/spf13/cobra"

var doMaterialTransactionAcceptancesCmd = &cobra.Command{
	Use:   "material-transaction-acceptances",
	Short: "Manage material transaction acceptances",
	Long: `Accept material transactions.

Acceptances move material transactions into the accepted status.

Commands:
  create    Accept a material transaction`,
	Example: `  # Accept a material transaction
  xbe do material-transaction-acceptances create --material-transaction 123 --comment "Reviewed"

  # Accept while skipping overlap validation
  xbe do material-transaction-acceptances create --material-transaction 123 --skip-not-overlapping-validation`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionAcceptancesCmd)
}
