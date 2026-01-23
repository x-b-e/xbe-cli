package cli

import "github.com/spf13/cobra"

var geofenceRestrictionsCmd = &cobra.Command{
	Use:   "geofence-restrictions",
	Short: "View geofence restrictions",
	Long: `View geofence restrictions.

Geofence restrictions define custom trucker access rules for geofences and
configure how often violation notifications can be sent.

Commands:
  list    List geofence restrictions with filtering
  show    Show geofence restriction details`,
	Example: `  # List geofence restrictions
  xbe view geofence-restrictions list

  # Filter by geofence
  xbe view geofence-restrictions list --geofence 123

  # Show a specific restriction
  xbe view geofence-restrictions show 456`,
}

func init() {
	viewCmd.AddCommand(geofenceRestrictionsCmd)
}
