package cli

import "github.com/spf13/cobra"

var doJobProductionPlanCancellationsCmd = &cobra.Command{
	Use:     "job-production-plan-cancellations",
	Aliases: []string{"job-production-plan-cancellation"},
	Short:   "Cancel job production plans",
	Long:    "Commands for canceling job production plans.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanCancellationsCmd)
}
