package cli

import "github.com/spf13/cobra"

var doJobProductionPlanRejectionsCmd = &cobra.Command{
	Use:     "job-production-plan-rejections",
	Aliases: []string{"job-production-plan-rejection"},
	Short:   "Reject job production plans",
	Long:    "Commands for rejecting job production plans.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanRejectionsCmd)
}
