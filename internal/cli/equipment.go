package cli

import "github.com/spf13/cobra"

var equipmentCmd = &cobra.Command{
	Use:     "equipment",
	Aliases: []string{"equip"},
	Short:   "View equipment",
	Long:    "Commands for viewing equipment.",
}

func init() {
	viewCmd.AddCommand(equipmentCmd)
}
