package cli

import "github.com/spf13/cobra"

var inventoryChangesCmd = &cobra.Command{
	Use:   "inventory-changes",
	Short: "Browse and view inventory changes",
	Long: `Browse and view inventory changes.

Inventory changes capture recalculations of material site inventories over time.
Each change records a forecast window, calculation timestamp, and the net impact
on inventory amounts.

Commands:
  list    List inventory changes with filtering
  show    View full inventory change details`,
	Example: `  # List inventory changes
  xbe view inventory-changes list

  # Filter by material site
  xbe view inventory-changes list --material-site 123

  # Filter by material type
  xbe view inventory-changes list --material-type 456

  # View a specific inventory change
  xbe view inventory-changes show 789`,
}

func init() {
	viewCmd.AddCommand(inventoryChangesCmd)
}
