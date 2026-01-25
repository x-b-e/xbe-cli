package cli

import "github.com/spf13/cobra"

var doEquipmentMovementTripDispatchesCmd = &cobra.Command{
	Use:     "equipment-movement-trip-dispatches",
	Aliases: []string{"equipment-movement-trip-dispatch"},
	Short:   "Manage equipment movement trip dispatches",
	Long:    "Commands for creating, updating, and deleting equipment movement trip dispatches.",
}

func init() {
	doCmd.AddCommand(doEquipmentMovementTripDispatchesCmd)
}
