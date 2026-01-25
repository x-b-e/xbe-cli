package cli

import "github.com/spf13/cobra"

var doJobProductionPlanScheduleChangesCmd = &cobra.Command{
	Use:     "job-production-plan-schedule-changes",
	Aliases: []string{"job-production-plan-schedule-change"},
	Short:   "Apply schedule changes to job production plans",
	Long:    "Commands for applying schedule changes to job production plans.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanScheduleChangesCmd)
}
