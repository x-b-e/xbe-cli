package cli

import "github.com/spf13/cobra"

var haskellLemonInboundMaterialTransactionExportsCmd = &cobra.Command{
	Use:     "haskell-lemon-inbound-material-transaction-exports",
	Aliases: []string{"haskell-lemon-inbound-material-transaction-export"},
	Short:   "Browse Haskell Lemon inbound material transaction exports",
	Long: `Browse Haskell Lemon inbound material transaction exports.

Exports capture inbound material transaction CSVs generated for Haskell Lemon
on a specific transaction date.

Commands:
  list    List exports with filtering
  show    Show export details`,
	Example: `  # List exports
  xbe view haskell-lemon-inbound-material-transaction-exports list

  # Show an export
  xbe view haskell-lemon-inbound-material-transaction-exports show 123`,
}

func init() {
	viewCmd.AddCommand(haskellLemonInboundMaterialTransactionExportsCmd)
}
