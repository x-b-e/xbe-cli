package cli

import "github.com/spf13/cobra"

var doJobProductionPlanAlarmsCmd = &cobra.Command{
	Use:   "job-production-plan-alarms",
	Short: "Manage job production plan alarms",
	Long: `Manage job production plan alarms.

Commands:
  create    Create a job production plan alarm
  update    Update a job production plan alarm
  delete    Delete a job production plan alarm`,
	Example: `  # Create an alarm
  xbe do job-production-plan-alarms create \
    --job-production-plan 123 \
    --tons 150 \
    --note "Notify at 150 tons"

  # Update an alarm
  xbe do job-production-plan-alarms update 456 --max-latency-minutes 60

  # Delete an alarm
  xbe do job-production-plan-alarms delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanAlarmsCmd)
}
