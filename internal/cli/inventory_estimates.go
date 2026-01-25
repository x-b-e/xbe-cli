package cli

import "github.com/spf13/cobra"

var inventoryEstimatesCmd = &cobra.Command{
	Use:     "inventory-estimates",
	Aliases: []string{"inventory-estimate"},
	Short:   "View inventory estimates",
	Long: `View inventory estimates.

Inventory estimates capture estimated inventory levels for material sites and
material types, including estimated amounts in tons and timestamps.

Commands:
  list    List inventory estimates with filtering
  show    Show inventory estimate details`,
}

func init() {
	viewCmd.AddCommand(inventoryEstimatesCmd)
}
