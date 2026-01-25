package cli

import "github.com/spf13/cobra"

var doInventoryCapacitiesCmd = &cobra.Command{
	Use:     "inventory-capacities",
	Aliases: []string{"inventory-capacity"},
	Short:   "Manage inventory capacities",
	Long:    "Create, update, and delete inventory capacities.",
}

func init() {
	doCmd.AddCommand(doInventoryCapacitiesCmd)
}
