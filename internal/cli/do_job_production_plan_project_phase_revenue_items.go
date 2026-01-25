package cli

import "github.com/spf13/cobra"

var doJobProductionPlanProjectPhaseRevenueItemsCmd = &cobra.Command{
	Use:     "job-production-plan-project-phase-revenue-items",
	Aliases: []string{"job-production-plan-project-phase-revenue-item"},
	Short:   "Manage job production plan project phase revenue items",
	Long:    "Commands for creating, updating, and deleting job production plan project phase revenue items.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanProjectPhaseRevenueItemsCmd)
}
