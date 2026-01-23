package cli

import "github.com/spf13/cobra"

var driverDayAdjustmentPlansCmd = &cobra.Command{
	Use:     "driver-day-adjustment-plans",
	Aliases: []string{"driver-day-adjustment-plan"},
	Short:   "View driver day adjustment plans",
	Long:    "Commands for viewing driver day adjustment plans.",
}

func init() {
	viewCmd.AddCommand(driverDayAdjustmentPlansCmd)
}
