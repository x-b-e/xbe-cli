package cli

import "github.com/spf13/cobra"

var transportOrderMaterialsCmd = &cobra.Command{
	Use:     "transport-order-materials",
	Aliases: []string{"transport-order-material"},
	Short:   "View transport order materials",
	Long: `View materials tied to transport orders.

Commands:
  list    List transport order materials with filters
  show    Show transport order material details`,
	Example: `  # List transport order materials
  xbe view transport-order-materials list

  # Filter by transport order
  xbe view transport-order-materials list --transport-order 123

  # Show a transport order material
  xbe view transport-order-materials show 456`,
}

func init() {
	viewCmd.AddCommand(transportOrderMaterialsCmd)
}
