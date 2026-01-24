package cli

import "github.com/spf13/cobra"

var projectPhaseRevenueItemsCmd = &cobra.Command{
	Use:     "project-phase-revenue-items",
	Aliases: []string{"project-phase-revenue-item"},
	Short:   "View project phase revenue items",
	Long:    "Commands for viewing project phase revenue items.",
}

func init() {
	viewCmd.AddCommand(projectPhaseRevenueItemsCmd)
}
