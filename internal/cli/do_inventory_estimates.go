package cli

import "github.com/spf13/cobra"

var doInventoryEstimatesCmd = &cobra.Command{
	Use:     "inventory-estimates",
	Aliases: []string{"inventory-estimate"},
	Short:   "Manage inventory estimates",
	Long:    "Commands for creating, updating, and deleting inventory estimates.",
}

func init() {
	doCmd.AddCommand(doInventoryEstimatesCmd)
}
