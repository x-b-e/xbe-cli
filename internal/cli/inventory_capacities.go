package cli

import "github.com/spf13/cobra"

var inventoryCapacitiesCmd = &cobra.Command{
	Use:     "inventory-capacities",
	Aliases: []string{"inventory-capacity"},
	Short:   "View inventory capacities",
	Long: `View inventory capacities on the XBE platform.

Inventory capacities define min/max storage levels and alert thresholds for
material sites and material types.

Commands:
  list    List inventory capacities with filtering
  show    Show inventory capacity details`,
	Example: `  # List inventory capacities
  xbe view inventory-capacities list

  # Filter by material site
  xbe view inventory-capacities list --material-site 123

  # Filter by material type
  xbe view inventory-capacities list --material-type 456

  # Show capacity details
  xbe view inventory-capacities show 789`,
}

func init() {
	viewCmd.AddCommand(inventoryCapacitiesCmd)
}
