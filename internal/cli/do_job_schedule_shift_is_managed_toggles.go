package cli

import "github.com/spf13/cobra"

var doJobScheduleShiftIsManagedTogglesCmd = &cobra.Command{
	Use:   "job-schedule-shift-is-managed-toggles",
	Short: "Toggle job schedule shift managed status",
	Long: `Toggle job schedule shift managed status.

Toggles the managed status for a job schedule shift and any related tender job
schedule shifts.

Commands:
  create    Toggle managed status for a job schedule shift`,
	Example: `  # Toggle managed status for a job schedule shift
  xbe do job-schedule-shift-is-managed-toggles create --job-schedule-shift 123`,
}

func init() {
	doCmd.AddCommand(doJobScheduleShiftIsManagedTogglesCmd)
}
