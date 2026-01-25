package cli

import "github.com/spf13/cobra"

var doTransportOrderStopMaterialsCmd = &cobra.Command{
	Use:   "transport-order-stop-materials",
	Short: "Manage transport order stop materials",
	Long: `Create, update, and delete transport order stop materials.

Transport order stop materials link a transport order material to a
specific stop, including the quantity at that stop.

Commands:
  create    Create a transport order stop material
  update    Update a transport order stop material
  delete    Delete a transport order stop material`,
}

func init() {
	doCmd.AddCommand(doTransportOrderStopMaterialsCmd)
}
