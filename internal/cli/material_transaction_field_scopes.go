package cli

import "github.com/spf13/cobra"

var materialTransactionFieldScopesCmd = &cobra.Command{
	Use:   "material-transaction-field-scopes",
	Short: "Browse material transaction field scopes",
	Long: `Browse material transaction field scopes on the XBE platform.

Material transaction field scopes surface matching context for a material
transaction, including ticket metadata and related job/material selections.

Commands:
  list  List material transaction field scopes (admin only)
  show  Show details for a material transaction field scope`,
	Example: `  # Show field scope details for a material transaction
  xbe view material-transaction-field-scopes show <material-transaction-id>

  # List field scopes (admin only)
  xbe view material-transaction-field-scopes list`,
}

func init() {
	viewCmd.AddCommand(materialTransactionFieldScopesCmd)
}
