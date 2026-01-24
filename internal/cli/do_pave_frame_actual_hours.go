package cli

import "github.com/spf13/cobra"

var doPaveFrameActualHoursCmd = &cobra.Command{
	Use:   "pave-frame-actual-hours",
	Short: "Manage pave frame actual hours",
	Long: `Create, update, and delete pave frame actual hours.

Note: Only admin users can create, update, or delete pave frame actual hours.`,
	Example: `  # Create a pave frame actual hour
  xbe do pave-frame-actual-hours create --date 2024-01-15 --hour 9 --window day \
    --latitude 38.9 --longitude -77.0 --temp-min-f 42.3 --precip-1hr-in 0.05

  # Update a record
  xbe do pave-frame-actual-hours update 123 --temp-min-f 45.0

  # Delete a record
  xbe do pave-frame-actual-hours delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doPaveFrameActualHoursCmd)
}
