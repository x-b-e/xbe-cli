package cli

import "github.com/spf13/cobra"

var transportRoutesCmd = &cobra.Command{
	Use:   "transport-routes",
	Short: "View transport routes",
	Long: `View transport routes on the XBE platform.

Transport routes represent computed paths between origin and destination
coordinates, including distance, duration, and polyline geometry.

Commands:
  list    List transport routes with filters
  show    Show transport route details`,
	Example: `  # List routes near an origin location
  xbe view transport-routes list --near-origin-location "40.7128|-74.0060|10"

  # List routes near a destination location
  xbe view transport-routes list --near-destination-location "34.0522|-118.2437|25"

  # Show a route by ID
  xbe view transport-routes show 123

  # JSON output
  xbe view transport-routes list --json`,
}

func init() {
	viewCmd.AddCommand(transportRoutesCmd)
}
