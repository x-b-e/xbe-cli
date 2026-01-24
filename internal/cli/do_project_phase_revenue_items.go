package cli

import "github.com/spf13/cobra"

var doProjectPhaseRevenueItemsCmd = &cobra.Command{
	Use:     "project-phase-revenue-items",
	Aliases: []string{"project-phase-revenue-item"},
	Short:   "Manage project phase revenue items",
	Long:    "Commands for creating, updating, and deleting project phase revenue items.",
}

func init() {
	doCmd.AddCommand(doProjectPhaseRevenueItemsCmd)
}
