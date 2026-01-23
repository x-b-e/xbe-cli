package cli

import "github.com/spf13/cobra"

var doEquipmentMovementStopRequirementsCmd = &cobra.Command{
	Use:     "equipment-movement-stop-requirements",
	Aliases: []string{"equipment-movement-stop-requirement"},
	Short:   "Manage equipment movement stop requirements",
	Long:    "Commands for creating and deleting equipment movement stop requirements.",
}

func init() {
	doCmd.AddCommand(doEquipmentMovementStopRequirementsCmd)
}
