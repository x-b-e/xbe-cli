package cli

import "github.com/spf13/cobra"

var jobProductionPlanAlarmSubscribersCmd = &cobra.Command{
	Use:   "job-production-plan-alarm-subscribers",
	Short: "View job production plan alarm subscribers",
	Long: `View job production plan alarm subscribers on the XBE platform.

Job production plan alarm subscribers define which users receive
notifications when an alarm is fulfilled.

Commands:
  list    List job production plan alarm subscribers
  show    Show job production plan alarm subscriber details`,
	Example: `  # List alarm subscribers
  xbe view job-production-plan-alarm-subscribers list

  # Show an alarm subscriber
  xbe view job-production-plan-alarm-subscribers show 123

  # Output as JSON
  xbe view job-production-plan-alarm-subscribers list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanAlarmSubscribersCmd)
}
