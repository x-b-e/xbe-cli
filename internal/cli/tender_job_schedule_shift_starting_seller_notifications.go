package cli

import "github.com/spf13/cobra"

var tenderJobScheduleShiftStartingSellerNotificationsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-starting-seller-notifications",
	Aliases: []string{"tender-job-schedule-shift-starting-seller-notification"},
	Short:   "Browse tender job schedule shift starting seller notifications",
	Long: `Browse tender job schedule shift starting seller notifications.

Tender job schedule shift starting seller notifications alert seller contacts
when a tender job schedule shift is about to start. Use list to filter and show
to view full details.`,
	Example: `  # List recent notifications
  xbe view tender-job-schedule-shift-starting-seller-notifications list

  # Filter by read status
  xbe view tender-job-schedule-shift-starting-seller-notifications list --read true

  # Show a notification
  xbe view tender-job-schedule-shift-starting-seller-notifications show 123`,
}

func init() {
	viewCmd.AddCommand(tenderJobScheduleShiftStartingSellerNotificationsCmd)
}
