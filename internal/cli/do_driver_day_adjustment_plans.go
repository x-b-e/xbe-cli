package cli

import "github.com/spf13/cobra"

var doDriverDayAdjustmentPlansCmd = &cobra.Command{
	Use:   "driver-day-adjustment-plans",
	Short: "Manage driver day adjustment plans",
	Long: `Manage driver day adjustment plans on the XBE platform.

Commands:
  create    Create a new driver day adjustment plan
  update    Update an existing driver day adjustment plan
  delete    Delete a driver day adjustment plan`,
	Example: `  # Create a plan
  xbe do driver-day-adjustment-plans create --trucker 123 --content "Updated schedule" --start-at "2025-01-15T08:00:00Z"

  # Update a plan
  xbe do driver-day-adjustment-plans update 456 --content "Adjusted for weather" --start-at "2025-01-16T06:00:00Z"

  # Delete a plan (requires --confirm)
  xbe do driver-day-adjustment-plans delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doDriverDayAdjustmentPlansCmd)
}
