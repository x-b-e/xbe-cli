package cli

import "github.com/spf13/cobra"

var doJobProductionPlanUncancellationsCmd = &cobra.Command{
	Use:   "job-production-plan-uncancellations",
	Short: "Uncancel job production plans",
	Long: `Uncancel job production plans.

Uncancelling a plan restores it to the previous status that was active
before it was cancelled.

Commands:
  create    Uncancel a job production plan`,
	Example: `  # Uncancel a job production plan
  xbe do job-production-plan-uncancellations create --job-production-plan 123

  # Uncancel with a comment and suppress notifications
  xbe do job-production-plan-uncancellations create \
    --job-production-plan 123 \
    --comment "Reopen plan" \
    --suppress-status-change-notifications`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanUncancellationsCmd)
}
