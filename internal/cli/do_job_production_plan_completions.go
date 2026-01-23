package cli

import "github.com/spf13/cobra"

var doJobProductionPlanCompletionsCmd = &cobra.Command{
	Use:   "job-production-plan-completions",
	Short: "Complete job production plans",
	Long: `Complete job production plans on the XBE platform.

Completions transition job production plans to complete status. Only plans
in approved status can be completed.

Commands:
  create    Complete a job production plan`,
	Example: `  # Complete a job production plan
  xbe do job-production-plan-completions create --job-production-plan 123 --comment "Finished"

  # Complete and suppress notifications
  xbe do job-production-plan-completions create --job-production-plan 123 --suppress-status-change-notifications`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanCompletionsCmd)
}
