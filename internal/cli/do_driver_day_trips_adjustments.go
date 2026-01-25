package cli

import "github.com/spf13/cobra"

var doDriverDayTripsAdjustmentsCmd = &cobra.Command{
	Use:     "driver-day-trips-adjustments",
	Aliases: []string{"driver-day-trips-adjustment"},
	Short:   "Manage driver day trips adjustments",
	Long:    "Commands for creating, updating, and deleting driver day trips adjustments.",
}

func init() {
	doCmd.AddCommand(doDriverDayTripsAdjustmentsCmd)
}
