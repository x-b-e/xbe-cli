package cli

import "github.com/spf13/cobra"

var doEquipmentMovementRequirementLocationsCmd = &cobra.Command{
	Use:     "equipment-movement-requirement-locations",
	Aliases: []string{"equipment-movement-requirement-location"},
	Short:   "Manage equipment movement requirement locations",
	Long:    "Commands for creating, updating, and deleting equipment movement requirement locations.",
}

func init() {
	doCmd.AddCommand(doEquipmentMovementRequirementLocationsCmd)
}
