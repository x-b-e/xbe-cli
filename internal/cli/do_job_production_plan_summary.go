package cli

import "github.com/spf13/cobra"

var doJobProductionPlanSummaryCmd = &cobra.Command{
	Use:   "job-production-plan-summary",
	Short: "Generate job production plan summaries",
	Long: `Generate job production plan summaries on the XBE platform.

Commands:
  create    Generate a job production plan summary`,
	Example: `  # Generate a job production plan summary grouped by customer
  xbe summarize job-production-plan-summary create --group-by customer --filter broker=123 --filter start_on=2025-01-01 --filter end_on=2025-01-31

  # Summary by project and date
  xbe summarize job-production-plan-summary create --group-by project,date --filter broker=123 --filter start_on=2025-01-01 --filter end_on=2025-01-31

  # Summary with specific metrics
  xbe summarize job-production-plan-summary create --group-by customer --filter broker=123 --filter start_on=2025-01-01 --filter end_on=2025-01-31 --metrics plan_count,tons_sum,truck_hours_sum`,
}

func init() {
	summarizeCmd.AddCommand(doJobProductionPlanSummaryCmd)
}
