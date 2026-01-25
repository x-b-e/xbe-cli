package cli

import "github.com/spf13/cobra"

var vehicleLocationEventsCmd = &cobra.Command{
	Use:     "vehicle-location-events",
	Aliases: []string{"vehicle-location-event"},
	Short:   "View vehicle location events",
	Long:    "Commands for viewing vehicle location events.",
}

func init() {
	viewCmd.AddCommand(vehicleLocationEventsCmd)
}
