package cli

import "github.com/spf13/cobra"

var doJobProductionPlanUncompletionsCmd = &cobra.Command{
	Use:   "job-production-plan-uncompletions",
	Short: "Uncomplete job production plans",
	Long: `Uncomplete job production plans.

Job production plan uncompletions transition a plan from complete to approved.

Commands:
  create    Uncomplete a job production plan`,
	Example: `  # Uncomplete a job production plan
  xbe do job-production-plan-uncompletions create --job-production-plan 12345`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanUncompletionsCmd)
}
