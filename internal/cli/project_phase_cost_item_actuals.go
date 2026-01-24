package cli

import "github.com/spf13/cobra"

var projectPhaseCostItemActualsCmd = &cobra.Command{
	Use:     "project-phase-cost-item-actuals",
	Aliases: []string{"project-phase-cost-item-actual"},
	Short:   "Browse project phase cost item actuals",
	Long:    "Commands for viewing project phase cost item actuals.",
}

func init() {
	viewCmd.AddCommand(projectPhaseCostItemActualsCmd)
}
