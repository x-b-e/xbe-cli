package cli

import "github.com/spf13/cobra"

var doMaterialPurchaseOrderReleasesCmd = &cobra.Command{
	Use:   "material-purchase-order-releases",
	Short: "Manage material purchase order releases",
	Long: `Create, update, and delete material purchase order releases.

Material purchase order releases allocate purchase order quantity to truckers
and shifts, tracking approval and redemption status.

Commands:
  create    Create a new release
  update    Update an existing release
  delete    Delete a release`,
}

func init() {
	doCmd.AddCommand(doMaterialPurchaseOrderReleasesCmd)
}
