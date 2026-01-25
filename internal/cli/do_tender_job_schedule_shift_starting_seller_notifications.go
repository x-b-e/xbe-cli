package cli

import "github.com/spf13/cobra"

var doTenderJobScheduleShiftStartingSellerNotificationsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-starting-seller-notifications",
	Aliases: []string{"tender-job-schedule-shift-starting-seller-notification"},
	Short:   "Manage tender job schedule shift starting seller notifications",
	Long:    "Commands for updating tender job schedule shift starting seller notifications.",
}

func init() {
	doCmd.AddCommand(doTenderJobScheduleShiftStartingSellerNotificationsCmd)
}
