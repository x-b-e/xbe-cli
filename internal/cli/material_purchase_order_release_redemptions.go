package cli

import "github.com/spf13/cobra"

var materialPurchaseOrderReleaseRedemptionsCmd = &cobra.Command{
	Use:   "material-purchase-order-release-redemptions",
	Short: "View material purchase order release redemptions",
	Long: `View material purchase order release redemptions on the XBE platform.

Material purchase order release redemptions link a purchase order release
with either a ticket number or a material transaction once redeemed.

Commands:
  list    List release redemptions with filtering
  show    Show release redemption details`,
	Example: `  # List release redemptions
  xbe view material-purchase-order-release-redemptions list

  # Filter by release
  xbe view material-purchase-order-release-redemptions list --release 123

  # Show a redemption
  xbe view material-purchase-order-release-redemptions show 456

  # Output as JSON
  xbe view material-purchase-order-release-redemptions list --json`,
}

func init() {
	viewCmd.AddCommand(materialPurchaseOrderReleaseRedemptionsCmd)
}
