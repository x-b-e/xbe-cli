package cli

import "github.com/spf13/cobra"

var driverMovementSegmentSetsCmd = &cobra.Command{
	Use:     "driver-movement-segment-sets",
	Aliases: []string{"driver-movement-segment-set"},
	Short:   "Browse driver movement segment sets",
	Long: `Browse driver movement segment sets.

Driver movement segment sets group movement segments for a driver day,
including summary metrics like total distance and moving time.

Commands:
  list    List driver movement segment sets with filtering and pagination
  show    Show full details of a driver movement segment set`,
	Example: `  # List driver movement segment sets
  xbe view driver-movement-segment-sets list

  # Filter by driver day
  xbe view driver-movement-segment-sets list --driver-day 123

  # Show a driver movement segment set
  xbe view driver-movement-segment-sets show 456`,
}

func init() {
	viewCmd.AddCommand(driverMovementSegmentSetsCmd)
}
