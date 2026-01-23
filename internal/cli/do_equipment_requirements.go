package cli

import "github.com/spf13/cobra"

var doEquipmentRequirementsCmd = &cobra.Command{
	Use:     "equipment-requirements",
	Aliases: []string{"equipment-requirement"},
	Short:   "Manage equipment requirements",
	Long:    "Commands for creating, updating, and deleting equipment requirements.",
}

func init() {
	doCmd.AddCommand(doEquipmentRequirementsCmd)
}
