package cli

import "github.com/spf13/cobra"

var doGeofenceRestrictionsCmd = &cobra.Command{
	Use:   "geofence-restrictions",
	Short: "Manage geofence restrictions",
	Long: `Create, update, and delete geofence restrictions.

Geofence restrictions assign specific truckers to geofences and configure
notification pacing for restriction violations.

Commands:
  create    Create a geofence restriction
  update    Update a geofence restriction
  delete    Delete a geofence restriction`,
	Example: `  # Create a geofence restriction
  xbe do geofence-restrictions create --geofence 123 --trucker 456

  # Update status and notification pacing
  xbe do geofence-restrictions update 789 --status inactive --max-seconds-between-notification 600

  # Delete a geofence restriction
  xbe do geofence-restrictions delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doGeofenceRestrictionsCmd)
}
