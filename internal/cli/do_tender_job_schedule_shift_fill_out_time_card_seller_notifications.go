package cli

import "github.com/spf13/cobra"

var doTenderJobScheduleShiftFillOutTimeCardSellerNotificationsCmd = &cobra.Command{
	Use:   "tender-job-schedule-shift-fill-out-time-card-seller-notifications",
	Short: "Manage tender job schedule shift fill out time card seller notifications",
	Long: `Commands for updating tender job schedule shift fill out time card seller notifications.

Notifications are created by the platform and can be marked as read by the
recipient.`,
	Example: `  # Mark a notification as read
  xbe do tender-job-schedule-shift-fill-out-time-card-seller-notifications update 123 --read`,
}

func init() {
	doCmd.AddCommand(doTenderJobScheduleShiftFillOutTimeCardSellerNotificationsCmd)
}
