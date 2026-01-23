package cli

import "github.com/spf13/cobra"

var transportOrderStopMaterialsCmd = &cobra.Command{
	Use:   "transport-order-stop-materials",
	Short: "View transport order stop materials",
	Long: `View transport order stop materials on the XBE platform.

Transport order stop materials link a transport order material to a
specific stop, capturing quantity at that stop.

Commands:
  list    List transport order stop materials
  show    Show transport order stop material details`,
	Example: `  # List transport order stop materials
  xbe view transport-order-stop-materials list

  # Show a specific transport order stop material
  xbe view transport-order-stop-materials show 123`,
}

func init() {
	viewCmd.AddCommand(transportOrderStopMaterialsCmd)
}
