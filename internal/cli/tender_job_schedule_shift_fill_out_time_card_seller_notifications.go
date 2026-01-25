package cli

import "github.com/spf13/cobra"

var tenderJobScheduleShiftFillOutTimeCardSellerNotificationsCmd = &cobra.Command{
	Use:   "tender-job-schedule-shift-fill-out-time-card-seller-notifications",
	Short: "View tender job schedule shift fill out time card seller notifications",
	Long: `View tender job schedule shift fill out time card seller notifications on the XBE platform.

These notifications are sent to sellers when a time card needs to be filled out.

Commands:
  list    List tender job schedule shift fill out time card seller notifications
  show    Show tender job schedule shift fill out time card seller notification details`,
	Example: `  # List notifications
  xbe view tender-job-schedule-shift-fill-out-time-card-seller-notifications list

  # Show a notification
  xbe view tender-job-schedule-shift-fill-out-time-card-seller-notifications show 123

  # Output as JSON
  xbe view tender-job-schedule-shift-fill-out-time-card-seller-notifications list --json`,
}

func init() {
	viewCmd.AddCommand(tenderJobScheduleShiftFillOutTimeCardSellerNotificationsCmd)
}
