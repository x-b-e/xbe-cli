package cli

import "github.com/spf13/cobra"

func newGeofencesShowCmd() *cobra.Command {
	return newGenericShowCmd("geofences")
}

func init() {
	geofencesCmd.AddCommand(newGeofencesShowCmd())
}
