package cli

import "github.com/spf13/cobra"

var doMaterialPurchaseOrderReleaseRedemptionsCmd = &cobra.Command{
	Use:   "material-purchase-order-release-redemptions",
	Short: "Manage material purchase order release redemptions",
	Long: `Create, update, and delete material purchase order release redemptions.

Redemptions bind a purchase order release to a ticket or material transaction.`,
}

func init() {
	doCmd.AddCommand(doMaterialPurchaseOrderReleaseRedemptionsCmd)
}
