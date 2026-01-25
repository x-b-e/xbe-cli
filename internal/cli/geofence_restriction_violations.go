package cli

import "github.com/spf13/cobra"

var geofenceRestrictionViolationsCmd = &cobra.Command{
	Use:     "geofence-restriction-violations",
	Aliases: []string{"geofence-restriction-violation"},
	Short:   "View geofence restriction violations",
	Long: `Commands for viewing geofence restriction violations.

Geofence restriction violations are recorded when a trailer, tractor, or driver
enters a geofence that has restrictions applied to them.`,
	Example: `  # List recent violations
  xbe view geofence-restriction-violations list --limit 25

  # Show a specific violation
  xbe view geofence-restriction-violations show 123`,
}

func init() {
	viewCmd.AddCommand(geofenceRestrictionViolationsCmd)
}
