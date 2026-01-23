package cli

import "github.com/spf13/cobra"

var jobProductionPlanSubscriptionsCmd = &cobra.Command{
	Use:   "job-production-plan-subscriptions",
	Short: "View job production plan subscriptions",
	Long: `View job production plan subscriptions on the XBE platform.

Job production plan subscriptions determine which users receive
notifications for a specific job production plan.

Commands:
  list    List job production plan subscriptions
  show    Show job production plan subscription details`,
	Example: `  # List subscriptions
  xbe view job-production-plan-subscriptions list

  # Show a subscription
  xbe view job-production-plan-subscriptions show 123

  # Output as JSON
  xbe view job-production-plan-subscriptions list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanSubscriptionsCmd)
}
