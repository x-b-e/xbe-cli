package cli

import "github.com/spf13/cobra"

var doHaskellLemonOutboundMaterialTransactionExportsCmd = &cobra.Command{
	Use:   "haskell-lemon-outbound-material-transaction-exports",
	Short: "Manage Haskell Lemon outbound material transaction exports",
	Long: `Manage Haskell Lemon outbound material transaction exports.

Exports generate CSVs of outbound material transactions for Haskell Lemon.

Commands:
  create    Create an outbound material transaction export`,
	Example: `  # Create an export for a transaction date
  xbe do haskell-lemon-outbound-material-transaction-exports create --transaction-date 2025-01-15`,
}

func init() {
	doCmd.AddCommand(doHaskellLemonOutboundMaterialTransactionExportsCmd)
}
