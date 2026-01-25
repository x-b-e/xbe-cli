package cli

import "github.com/spf13/cobra"

var jobProductionPlanScheduleChangeWorksCmd = &cobra.Command{
	Use:     "job-production-plan-schedule-change-works",
	Aliases: []string{"job-production-plan-schedule-change-work"},
	Short:   "Browse job production plan schedule change works",
	Long: `Browse job production plan schedule change works.

Schedule change works capture the background work that applies schedule changes to
job production plans.

Commands:
  list    List schedule change works with filtering and pagination
  show    Show full details of a schedule change work item`,
	Example: `  # List schedule change works
  xbe view job-production-plan-schedule-change-works list

  # Show schedule change work details
  xbe view job-production-plan-schedule-change-works show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanScheduleChangeWorksCmd)
}
