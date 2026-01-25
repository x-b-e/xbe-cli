package cli

import "github.com/spf13/cobra"

var equipmentLocationEventsCmd = &cobra.Command{
	Use:     "equipment-location-events",
	Aliases: []string{"equipment-location-event"},
	Short:   "View equipment location events",
	Long:    "Commands for viewing equipment location events.",
}

func init() {
	viewCmd.AddCommand(equipmentLocationEventsCmd)
}
