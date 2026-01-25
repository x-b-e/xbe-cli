package cli

import "github.com/spf13/cobra"

var doJobProductionPlanApprovalsCmd = &cobra.Command{
	Use:   "job-production-plan-approvals",
	Short: "Approve job production plans",
	Long: `Approve job production plans.

Job production plan approvals transition a plan from submitted to approved.

Commands:
  create    Approve a job production plan`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanApprovalsCmd)
}
