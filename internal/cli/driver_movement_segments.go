package cli

import "github.com/spf13/cobra"

var driverMovementSegmentsCmd = &cobra.Command{
	Use:     "driver-movement-segments",
	Aliases: []string{"driver-movement-segment"},
	Short:   "Browse driver movement segments",
	Long: `Browse driver movement segments.

Driver movement segments represent contiguous moving or stationary intervals
for a driver day, including timestamps, distance travelled, and optional
site context.

Commands:
  list    List movement segments with filtering and pagination
  show    Show full details of a movement segment`,
	Example: `  # List movement segments
  xbe view driver-movement-segments list

  # Filter moving segments
  xbe view driver-movement-segments list --is-moving true

  # Filter by segment set
  xbe view driver-movement-segments list --driver-movement-segment-set 123

  # Show segment details
  xbe view driver-movement-segments show 456`,
}

func init() {
	viewCmd.AddCommand(driverMovementSegmentsCmd)
}
