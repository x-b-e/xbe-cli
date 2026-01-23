package cli

import "github.com/spf13/cobra"

var doDriverDayAdjustmentsCmd = &cobra.Command{
	Use:   "driver-day-adjustments",
	Short: "Manage driver day adjustments",
	Long: `Manage driver day adjustments.

Commands:
  create    Create a driver day adjustment
  update    Update a driver day adjustment
  delete    Delete a driver day adjustment`,
	Example: `  # Create an adjustment with an explicit amount
  xbe do driver-day-adjustments create --driver-day 123 --amount-explicit "25.00"

  # Update the explicit amount
  xbe do driver-day-adjustments update 456 --amount-explicit "15.00"

  # Delete an adjustment
  xbe do driver-day-adjustments delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doDriverDayAdjustmentsCmd)
}
