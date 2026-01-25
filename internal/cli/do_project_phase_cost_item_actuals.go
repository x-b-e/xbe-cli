package cli

import "github.com/spf13/cobra"

var doProjectPhaseCostItemActualsCmd = &cobra.Command{
	Use:     "project-phase-cost-item-actuals",
	Aliases: []string{"project-phase-cost-item-actual"},
	Short:   "Manage project phase cost item actuals",
	Long:    "Commands for creating, updating, and deleting project phase cost item actuals.",
}

func init() {
	doCmd.AddCommand(doProjectPhaseCostItemActualsCmd)
}
