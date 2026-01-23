package cli

import "github.com/spf13/cobra"

var doDriverDayShortfallCalculationsCmd = &cobra.Command{
	Use:   "driver-day-shortfall-calculations",
	Short: "Calculate driver day shortfall allocations",
	Long: `Calculate driver day shortfall allocations on the XBE platform.

Commands:
  create    Calculate a driver day shortfall allocation`,
	Example: `  # Calculate a shortfall for specific time cards and constraints
  xbe do driver-day-shortfall-calculations create --time-card-ids 101,102 --driver-day-time-card-constraint-ids 55,56`,
}

func init() {
	doCmd.AddCommand(doDriverDayShortfallCalculationsCmd)
}
