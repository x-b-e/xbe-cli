package cli

import "github.com/spf13/cobra"

var tenderJobScheduleShiftCancellationsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-cancellations",
	Aliases: []string{"tender-job-schedule-shift-cancellation"},
	Short:   "View tender job schedule shift cancellations",
	Long: `View tender job schedule shift cancellations on the XBE platform.

Tender job schedule shift cancellations capture cancellation requests for tender job schedule shifts.

Commands:
  list    List tender job schedule shift cancellations
  show    Show tender job schedule shift cancellation details`,
	Example: `  # List cancellations
  xbe view tender-job-schedule-shift-cancellations list

  # Show a cancellation
  xbe view tender-job-schedule-shift-cancellations show 123`,
}

func init() {
	viewCmd.AddCommand(tenderJobScheduleShiftCancellationsCmd)
}
