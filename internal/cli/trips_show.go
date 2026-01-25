package cli

import "github.com/spf13/cobra"

func newTripsShowCmd() *cobra.Command {
	return newGenericShowCmd("trips")
}

func init() {
	tripsCmd.AddCommand(newTripsShowCmd())
}
