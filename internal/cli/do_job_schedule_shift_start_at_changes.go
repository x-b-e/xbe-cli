package cli

import "github.com/spf13/cobra"

var doJobScheduleShiftStartAtChangesCmd = &cobra.Command{
	Use:     "job-schedule-shift-start-at-changes",
	Aliases: []string{"job-schedule-shift-start-at-change"},
	Short:   "Manage job schedule shift start-at changes",
	Long: `Manage job schedule shift start-at changes.

Start-at changes reschedule job schedule shifts by updating their start times.

Commands:
  create  Create a start-at change`,
	Example: `  # Create a start-at change
  xbe do job-schedule-shift-start-at-changes create \\
    --job-schedule-shift 123 \\
    --new-start-at 2026-01-23T14:30:00Z`,
}

func init() {
	doCmd.AddCommand(doJobScheduleShiftStartAtChangesCmd)
}
