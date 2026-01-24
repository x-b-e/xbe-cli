package cli

import "github.com/spf13/cobra"

var doProjectPhaseRevenueItemActualsCmd = &cobra.Command{
	Use:     "project-phase-revenue-item-actuals",
	Aliases: []string{"project-phase-revenue-item-actual"},
	Short:   "Manage project phase revenue item actuals",
	Long:    "Commands for creating, updating, and deleting project phase revenue item actuals.",
}

func init() {
	doCmd.AddCommand(doProjectPhaseRevenueItemActualsCmd)
}
