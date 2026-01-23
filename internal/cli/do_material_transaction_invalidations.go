package cli

import "github.com/spf13/cobra"

var doMaterialTransactionInvalidationsCmd = &cobra.Command{
	Use:   "material-transaction-invalidations",
	Short: "Manage material transaction invalidations",
	Long: `Invalidate material transactions.

Invalidations move material transactions into the invalidated status.

Commands:
  create    Invalidate a material transaction`,
	Example: `  # Invalidate a material transaction
  xbe do material-transaction-invalidations create --material-transaction 123 --comment "Duplicate ticket"`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionInvalidationsCmd)
}
