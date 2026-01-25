package cli

import "github.com/spf13/cobra"

var doEquipmentLocationEventsCmd = &cobra.Command{
	Use:     "equipment-location-events",
	Aliases: []string{"equipment-location-event"},
	Short:   "Manage equipment location events",
	Long:    "Commands for creating, updating, and deleting equipment location events.",
}

func init() {
	doCmd.AddCommand(doEquipmentLocationEventsCmd)
}
