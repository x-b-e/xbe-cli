package cli

import "github.com/spf13/cobra"

var jobProductionPlanJobSiteLocationEstimatesCmd = &cobra.Command{
	Use:     "job-production-plan-job-site-location-estimates",
	Aliases: []string{"job-production-plan-job-site-location-estimate"},
	Short:   "View job production plan job site location estimates",
	Long: `View job production plan job site location estimates.

Job site location estimates capture calculated destination locations over time
for a job production plan.

Commands:
  list    List job production plan job site location estimates
  show    Show job production plan job site location estimate details`,
	Example: `  # List job site location estimates
  xbe view job-production-plan-job-site-location-estimates list

  # Show a job site location estimate
  xbe view job-production-plan-job-site-location-estimates show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanJobSiteLocationEstimatesCmd)
}
