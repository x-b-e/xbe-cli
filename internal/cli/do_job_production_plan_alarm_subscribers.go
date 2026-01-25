package cli

import "github.com/spf13/cobra"

var doJobProductionPlanAlarmSubscribersCmd = &cobra.Command{
	Use:   "job-production-plan-alarm-subscribers",
	Short: "Manage job production plan alarm subscribers",
	Long:  "Commands for creating and deleting job production plan alarm subscribers.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanAlarmSubscribersCmd)
}
