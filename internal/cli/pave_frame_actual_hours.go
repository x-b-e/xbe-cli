package cli

import "github.com/spf13/cobra"

var paveFrameActualHoursCmd = &cobra.Command{
	Use:   "pave-frame-actual-hours",
	Short: "View pave frame actual hours",
	Long: `View pave frame actual hours on the XBE platform.

Pave frame actual hours capture hourly paving conditions for specific
coordinates, including temperature and precipitation.

Commands:
  list    List pave frame actual hours
  show    Show pave frame actual hour details`,
	Example: `  # List pave frame actual hours
  xbe view pave-frame-actual-hours list

  # Show a specific record
  xbe view pave-frame-actual-hours show 123

  # Output as JSON
  xbe view pave-frame-actual-hours list --json`,
}

func init() {
	viewCmd.AddCommand(paveFrameActualHoursCmd)
}
