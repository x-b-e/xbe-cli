package cli

import "github.com/spf13/cobra"

var tenderJobScheduleShiftCancelledTruckerContactNotificationsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-cancelled-trucker-contact-notifications",
	Aliases: []string{"tender-job-schedule-shift-cancelled-trucker-contact-notification"},
	Short:   "Browse tender job schedule shift cancelled trucker contact notifications",
	Long: `Browse tender job schedule shift cancelled trucker contact notifications.

Tender job schedule shift cancelled trucker contact notifications alert trucker
contacts when an accepted broker tender shift is cancelled. Use list to filter
and show to view full details.`,
	Example: `  # List recent notifications
  xbe view tender-job-schedule-shift-cancelled-trucker-contact-notifications list

  # Filter by read status
  xbe view tender-job-schedule-shift-cancelled-trucker-contact-notifications list --read true

  # Show a notification
  xbe view tender-job-schedule-shift-cancelled-trucker-contact-notifications show 123`,
}

func init() {
	viewCmd.AddCommand(tenderJobScheduleShiftCancelledTruckerContactNotificationsCmd)
}
