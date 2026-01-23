package cli

import "github.com/spf13/cobra"

var jobProductionPlanProjectPhaseRevenueItemsCmd = &cobra.Command{
	Use:     "job-production-plan-project-phase-revenue-items",
	Aliases: []string{"job-production-plan-project-phase-revenue-item"},
	Short:   "Browse job production plan project phase revenue items",
	Long: `Browse job production plan project phase revenue items.

Job production plan project phase revenue items connect a job production plan
with a project phase revenue item and track quantity at the plan level.

Commands:
  list    List items with filters
  show    Show item details`,
	Example: `  # List items
  xbe view job-production-plan-project-phase-revenue-items list

  # Show details for an item
  xbe view job-production-plan-project-phase-revenue-items show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanProjectPhaseRevenueItemsCmd)
}
