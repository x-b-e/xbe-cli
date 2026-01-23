package cli

import "github.com/spf13/cobra"

var jobScheduleShiftStartAtChangesCmd = &cobra.Command{
	Use:     "job-schedule-shift-start-at-changes",
	Aliases: []string{"job-schedule-shift-start-at-change"},
	Short:   "Browse job schedule shift start-at changes",
	Long: `Browse job schedule shift start-at changes.

Start-at changes capture rescheduling events for job schedule shifts.

Commands:
  list    List start-at changes with filtering and pagination
  show    Show full details of a start-at change`,
	Example: `  # List start-at changes
  xbe view job-schedule-shift-start-at-changes list

  # Show start-at change details
  xbe view job-schedule-shift-start-at-changes show 123`,
}

func init() {
	viewCmd.AddCommand(jobScheduleShiftStartAtChangesCmd)
}
