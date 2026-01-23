package cli

import "github.com/spf13/cobra"

var jobProductionPlanTimeCardApproversCmd = &cobra.Command{
	Use:     "job-production-plan-time-card-approvers",
	Aliases: []string{"job-production-plan-time-card-approver"},
	Short:   "View job production plan time card approvers",
	Long: `View job production plan time card approvers.

Job production plan time card approvers link users who can approve
submitted time cards for a job production plan.

Commands:
  list    List job production plan time card approvers
  show    Show job production plan time card approver details`,
	Example: `  # List time card approvers
  xbe view job-production-plan-time-card-approvers list

  # Show a specific time card approver
  xbe view job-production-plan-time-card-approvers show 123

  # Output JSON
  xbe view job-production-plan-time-card-approvers list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanTimeCardApproversCmd)
}
