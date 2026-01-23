package cli

import "github.com/spf13/cobra"

var doEquipmentMovementTripsCmd = &cobra.Command{
	Use:     "equipment-movement-trips",
	Aliases: []string{"equipment-movement-trip"},
	Short:   "Manage equipment movement trips",
	Long:    "Commands for creating, updating, and deleting equipment movement trips.",
}

func init() {
	doCmd.AddCommand(doEquipmentMovementTripsCmd)
}
