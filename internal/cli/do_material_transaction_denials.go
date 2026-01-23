package cli

import "github.com/spf13/cobra"

var doMaterialTransactionDenialsCmd = &cobra.Command{
	Use:   "material-transaction-denials",
	Short: "Deny material transactions",
	Long: `Deny material transactions.

Denials set a material transaction's status to denied.

Commands:
  create    Deny a material transaction`,
	Example: `  # Deny a material transaction
  xbe do material-transaction-denials create \
    --material-transaction 123 \
    --comment "Load contaminated"`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionDenialsCmd)
}
