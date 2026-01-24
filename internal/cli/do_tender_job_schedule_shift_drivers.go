package cli

import "github.com/spf13/cobra"

var doTenderJobScheduleShiftDriversCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-drivers",
	Aliases: []string{"tender-job-schedule-shift-driver"},
	Short:   "Manage tender job schedule shift drivers",
	Long: `Commands for managing tender job schedule shift drivers.

Create, update, or delete driver assignments on tender job schedule shifts.`,
}

func init() {
	doCmd.AddCommand(doTenderJobScheduleShiftDriversCmd)
}
