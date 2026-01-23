package cli

import "github.com/spf13/cobra"

var driverDayAdjustmentsCmd = &cobra.Command{
	Use:     "driver-day-adjustments",
	Aliases: []string{"driver-day-adjustment"},
	Short:   "Browse driver day adjustments",
	Long: `Browse driver day adjustments.

Driver day adjustments capture explicit or plan-generated amounts applied to a driver day.

Commands:
  list    List adjustments with filtering and pagination
  show    Show full details of an adjustment`,
	Example: `  # List driver day adjustments
  xbe view driver-day-adjustments list

  # Filter by driver day
  xbe view driver-day-adjustments list --driver-day 123

  # Filter by trucker
  xbe view driver-day-adjustments list --trucker 456

  # Filter by driver
  xbe view driver-day-adjustments list --driver 789

  # Show an adjustment
  xbe view driver-day-adjustments show 321`,
}

func init() {
	viewCmd.AddCommand(driverDayAdjustmentsCmd)
}
