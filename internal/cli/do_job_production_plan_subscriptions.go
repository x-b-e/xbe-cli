package cli

import "github.com/spf13/cobra"

var doJobProductionPlanSubscriptionsCmd = &cobra.Command{
	Use:   "job-production-plan-subscriptions",
	Short: "Manage job production plan subscriptions",
	Long:  "Commands for creating, updating, and deleting job production plan subscriptions.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanSubscriptionsCmd)
}
