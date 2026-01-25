package cli

import "github.com/spf13/cobra"

var doJobProductionPlanTimeCardApproversCmd = &cobra.Command{
	Use:     "job-production-plan-time-card-approvers",
	Aliases: []string{"job-production-plan-time-card-approver"},
	Short:   "Manage job production plan time card approvers",
	Long: `Create and delete job production plan time card approvers.

Commands:
  create    Create a time card approver
  delete    Delete a time card approver`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanTimeCardApproversCmd)
}
