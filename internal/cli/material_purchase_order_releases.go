package cli

import "github.com/spf13/cobra"

var materialPurchaseOrderReleasesCmd = &cobra.Command{
	Use:   "material-purchase-order-releases",
	Short: "View material purchase order releases",
	Long: `Browse material purchase order releases on the XBE platform.

Material purchase order releases allocate a portion of a purchase order
quantity to a specific trucker or shift and track release status.

Commands:
  list    List material purchase order releases
  show    Show material purchase order release details`,
	Example: `  # List releases
  xbe view material-purchase-order-releases list

  # Show a release
  xbe view material-purchase-order-releases show 123`,
}

func init() {
	viewCmd.AddCommand(materialPurchaseOrderReleasesCmd)
}
