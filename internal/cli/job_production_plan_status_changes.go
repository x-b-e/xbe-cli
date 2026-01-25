package cli

import "github.com/spf13/cobra"

var jobProductionPlanStatusChangesCmd = &cobra.Command{
	Use:   "job-production-plan-status-changes",
	Short: "Browse job production plan status changes",
	Long: `Browse job production plan status changes on the XBE platform.

Status changes capture transitions of a job production plan's lifecycle status,
including who changed it and when.

Commands:
  list    List job production plan status changes
  show    Show job production plan status change details`,
	Example: `  # List recent status changes
  xbe view job-production-plan-status-changes list

  # Show status change details
  xbe view job-production-plan-status-changes show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanStatusChangesCmd)
}
