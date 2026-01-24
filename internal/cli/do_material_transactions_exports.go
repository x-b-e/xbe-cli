package cli

import "github.com/spf13/cobra"

var doMaterialTransactionsExportsCmd = &cobra.Command{
	Use:   "material-transactions-exports",
	Short: "Manage material transaction exports",
	Long: `Manage material transaction exports.

Exports generate formatted files for selected material transactions using an
organization formatter.

Commands:
  create    Create a material transaction export`,
	Example: `  # Create an export
  xbe do material-transactions-exports create --organization-formatter 123 --material-transaction-ids 456`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionsExportsCmd)
}
