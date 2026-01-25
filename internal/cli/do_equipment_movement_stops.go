package cli

import "github.com/spf13/cobra"

var doEquipmentMovementStopsCmd = &cobra.Command{
	Use:     "equipment-movement-stops",
	Aliases: []string{"equipment-movement-stop"},
	Short:   "Manage equipment movement stops",
	Long:    "Commands for creating, updating, and deleting equipment movement stops.",
}

func init() {
	doCmd.AddCommand(doEquipmentMovementStopsCmd)
}
