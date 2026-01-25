package cli

import "github.com/spf13/cobra"

var doHaskellLemonInboundMaterialTransactionExportsCmd = &cobra.Command{
	Use:     "haskell-lemon-inbound-material-transaction-exports",
	Aliases: []string{"haskell-lemon-inbound-material-transaction-export"},
	Short:   "Generate Haskell Lemon inbound material transaction exports",
	Long: `Create Haskell Lemon inbound material transaction export CSVs.

Exports are generated asynchronously and can email the CSV to configured
recipients unless created as a test export.

Commands:
  create  Create an inbound material transaction export`,
	Example: `  # Create an export for a transaction date
  xbe do haskell-lemon-inbound-material-transaction-exports create --transaction-date 2025-01-15

  # Create a test export with explicit recipients
  xbe do haskell-lemon-inbound-material-transaction-exports create \
    --transaction-date 2025-01-15 \
    --is-test \
    --to-addresses "ops@example.com"`,
}

func init() {
	doCmd.AddCommand(doHaskellLemonInboundMaterialTransactionExportsCmd)
}
