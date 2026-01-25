package cli

import "github.com/spf13/cobra"

var doInventoryChangesCmd = &cobra.Command{
	Use:   "inventory-changes",
	Short: "Manage inventory changes",
	Long: `Create and delete inventory changes.

Inventory changes recalculate inventory estimates for material sites and types.
Use create to trigger a recalculation for a specific time window.`,
}

func init() {
	doCmd.AddCommand(doInventoryChangesCmd)
}
