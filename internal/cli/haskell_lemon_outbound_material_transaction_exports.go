package cli

import "github.com/spf13/cobra"

var haskellLemonOutboundMaterialTransactionExportsCmd = &cobra.Command{
	Use:     "haskell-lemon-outbound-material-transaction-exports",
	Aliases: []string{"haskell-lemon-outbound-material-transaction-export"},
	Short:   "Browse Haskell Lemon outbound material transaction exports",
	Long: `Browse Haskell Lemon outbound material transaction exports.

These exports generate CSVs of outbound material transactions for the
Haskell Lemon branch.

Commands:
  list    List outbound material transaction exports
  show    Show outbound material transaction export details`,
	Example: `  # List exports
  xbe view haskell-lemon-outbound-material-transaction-exports list

  # Show export details
  xbe view haskell-lemon-outbound-material-transaction-exports show 123`,
}

func init() {
	viewCmd.AddCommand(haskellLemonOutboundMaterialTransactionExportsCmd)
}
