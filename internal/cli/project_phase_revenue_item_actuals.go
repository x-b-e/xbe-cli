package cli

import "github.com/spf13/cobra"

var projectPhaseRevenueItemActualsCmd = &cobra.Command{
	Use:     "project-phase-revenue-item-actuals",
	Aliases: []string{"project-phase-revenue-item-actual"},
	Short:   "Browse project phase revenue item actuals",
	Long:    "Commands for viewing project phase revenue item actuals.",
}

func init() {
	viewCmd.AddCommand(projectPhaseRevenueItemActualsCmd)
}
