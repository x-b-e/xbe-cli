package cli

import "github.com/spf13/cobra"

var doEquipmentMovementRequirementsCmd = &cobra.Command{
	Use:     "equipment-movement-requirements",
	Aliases: []string{"equipment-movement-requirement"},
	Short:   "Manage equipment movement requirements",
	Long:    "Commands for creating, updating, and deleting equipment movement requirements.",
}

func init() {
	doCmd.AddCommand(doEquipmentMovementRequirementsCmd)
}
