package cli

import "github.com/spf13/cobra"

var equipmentMovementRequirementsCmd = &cobra.Command{
	Use:     "equipment-movement-requirements",
	Aliases: []string{"equipment-movement-requirement"},
	Short:   "View equipment movement requirements",
	Long:    "Commands for viewing equipment movement requirements.",
}

func init() {
	viewCmd.AddCommand(equipmentMovementRequirementsCmd)
}
