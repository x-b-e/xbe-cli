package cli

import "github.com/spf13/cobra"

var driverDayTripsAdjustmentsCmd = &cobra.Command{
	Use:     "driver-day-trips-adjustments",
	Aliases: []string{"driver-day-trips-adjustment"},
	Short:   "Browse driver day trips adjustments",
	Long: `Browse driver day trips adjustments.

Driver day trips adjustments capture edits to a driver day's trip sequence
for a specific tender job schedule shift.

Commands:
  list    List adjustments with filtering and pagination
  show    Show full adjustment details`,
	Example: `  # List adjustments
  xbe view driver-day-trips-adjustments list

  # Filter by tender job schedule shift
  xbe view driver-day-trips-adjustments list --tender-job-schedule-shift 123

  # Show an adjustment
  xbe view driver-day-trips-adjustments show 456`,
}

func init() {
	viewCmd.AddCommand(driverDayTripsAdjustmentsCmd)
}
