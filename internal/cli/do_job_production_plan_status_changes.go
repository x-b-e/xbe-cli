package cli

import "github.com/spf13/cobra"

var doJobProductionPlanStatusChangesCmd = &cobra.Command{
	Use:   "job-production-plan-status-changes",
	Short: "Manage job production plan status changes",
	Long: `Update job production plan status changes, such as setting a cancellation
reason type for cancelled or scrapped plans.

Commands:
  update    Update a job production plan status change`,
	Example: `  # Set a cancellation reason type
  xbe do job-production-plan-status-changes update 123 --job-production-plan-cancellation-reason-type 456`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanStatusChangesCmd)
}
