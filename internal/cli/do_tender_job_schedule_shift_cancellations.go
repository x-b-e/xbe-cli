package cli

import "github.com/spf13/cobra"

var doTenderJobScheduleShiftCancellationsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-cancellations",
	Aliases: []string{"tender-job-schedule-shift-cancellation"},
	Short:   "Cancel tender job schedule shifts",
	Long: `Cancel tender job schedule shifts.

Shift cancellations cancel a single tender job schedule shift and optionally skip notifications.

Commands:
  create    Cancel a tender job schedule shift`,
	Example: `  # Cancel a tender job schedule shift
  xbe do tender-job-schedule-shift-cancellations create --tender-job-schedule-shift 123`,
}

func init() {
	doCmd.AddCommand(doTenderJobScheduleShiftCancellationsCmd)
}
