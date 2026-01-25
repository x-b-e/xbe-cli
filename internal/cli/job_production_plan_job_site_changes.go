package cli

import "github.com/spf13/cobra"

var jobProductionPlanJobSiteChangesCmd = &cobra.Command{
	Use:     "job-production-plan-job-site-changes",
	Aliases: []string{"job-production-plan-job-site-change"},
	Short:   "Browse job production plan job site changes",
	Long: `Browse job production plan job site changes.

Job site changes capture when a job production plan's job site is updated.

Commands:
  show    Show full details of a job site change`,
	Example: `  # Show job site change details
  xbe view job-production-plan-job-site-changes show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanJobSiteChangesCmd)
}
