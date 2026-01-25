package cli

import "github.com/spf13/cobra"

var doTenderJobScheduleShiftCancelledTruckerContactNotificationsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-cancelled-trucker-contact-notifications",
	Aliases: []string{"tender-job-schedule-shift-cancelled-trucker-contact-notification"},
	Short:   "Manage tender job schedule shift cancelled trucker contact notifications",
	Long:    "Commands for updating tender job schedule shift cancelled trucker contact notifications.",
}

func init() {
	doCmd.AddCommand(doTenderJobScheduleShiftCancelledTruckerContactNotificationsCmd)
}
