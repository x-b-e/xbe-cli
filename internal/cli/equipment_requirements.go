package cli

import "github.com/spf13/cobra"

var equipmentRequirementsCmd = &cobra.Command{
	Use:     "equipment-requirements",
	Aliases: []string{"equipment-requirement"},
	Short:   "View equipment requirements",
	Long:    "Commands for viewing equipment requirements.",
}

func init() {
	viewCmd.AddCommand(equipmentRequirementsCmd)
}
