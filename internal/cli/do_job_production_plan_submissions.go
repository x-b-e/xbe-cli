package cli

import "github.com/spf13/cobra"

var doJobProductionPlanSubmissionsCmd = &cobra.Command{
	Use:   "job-production-plan-submissions",
	Short: "Submit job production plans",
	Long: `Submit job production plans.

Submitting a plan moves it to the submitted status and runs submission validations.

Commands:
  create    Submit a job production plan`,
	Example: `  # Submit a job production plan
  xbe do job-production-plan-submissions create --job-production-plan 123

  # Submit with a comment and suppress notifications
  xbe do job-production-plan-submissions create \
    --job-production-plan 123 \
    --comment "Ready for review" \
    --suppress-status-change-notifications`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanSubmissionsCmd)
}
