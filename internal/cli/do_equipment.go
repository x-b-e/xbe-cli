package cli

import "github.com/spf13/cobra"

var doEquipmentCmd = &cobra.Command{
	Use:     "equipment",
	Aliases: []string{"equip"},
	Short:   "Manage equipment",
	Long:    "Commands for creating, updating, and deleting equipment.",
}

func init() {
	doCmd.AddCommand(doEquipmentCmd)
}
