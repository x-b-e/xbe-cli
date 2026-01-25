package cli

import "github.com/spf13/cobra"

var geofencesCmd = &cobra.Command{
	Use:     "geofences",
	Aliases: []string{"geofence"},
	Short:   "View geofences",
	Long:    "Commands for viewing geofences (geographic boundaries).",
}

func init() {
	viewCmd.AddCommand(geofencesCmd)
}
