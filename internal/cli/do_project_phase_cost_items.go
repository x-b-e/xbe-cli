package cli

import "github.com/spf13/cobra"

var doProjectPhaseCostItemsCmd = &cobra.Command{
	Use:     "project-phase-cost-items",
	Aliases: []string{"project-phase-cost-item"},
	Short:   "Manage project phase cost items",
	Long:    "Commands for creating, updating, and deleting project phase cost items.",
}

func init() {
	doCmd.AddCommand(doProjectPhaseCostItemsCmd)
}
