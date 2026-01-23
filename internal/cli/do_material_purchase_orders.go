package cli

import "github.com/spf13/cobra"

var doMaterialPurchaseOrdersCmd = &cobra.Command{
	Use:     "material-purchase-orders",
	Aliases: []string{"material-purchase-order"},
	Short:   "Manage material purchase orders",
	Long:    `Create, update, and delete material purchase orders.`,
}

func init() {
	doCmd.AddCommand(doMaterialPurchaseOrdersCmd)
}
