package cli

import "github.com/spf13/cobra"

var jobProductionPlanAlarmsCmd = &cobra.Command{
	Use:   "job-production-plan-alarms",
	Short: "Browse job production plan alarms",
	Long: `Browse job production plan alarms on the XBE platform.

Job production plan alarms notify subscribers when production reaches
specified tonnage thresholds or exceeds latency targets.

Commands:
  list    List job production plan alarms with filtering and pagination
  show    Show job production plan alarm details`,
	Example: `  # List alarms
  xbe view job-production-plan-alarms list

  # Show an alarm
  xbe view job-production-plan-alarms show 123

  # Output as JSON
  xbe view job-production-plan-alarms list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanAlarmsCmd)
}
