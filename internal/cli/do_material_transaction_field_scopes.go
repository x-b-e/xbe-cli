package cli

import "github.com/spf13/cobra"

var doMaterialTransactionFieldScopesCmd = &cobra.Command{
	Use:   "material-transaction-field-scopes",
	Short: "Manage material transaction field scopes",
	Long: `Create material transaction field scopes for matching diagnostics.

Commands:
  create  Create a material transaction field scope`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionFieldScopesCmd)
}
