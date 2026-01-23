package cli

import "github.com/spf13/cobra"

var doGeofencesCmd = &cobra.Command{
	Use:     "geofences",
	Aliases: []string{"geofence"},
	Short:   "Manage geofences",
	Long:    "Commands for creating, updating, and deleting geofences.",
}

func init() {
	doCmd.AddCommand(doGeofencesCmd)
}
