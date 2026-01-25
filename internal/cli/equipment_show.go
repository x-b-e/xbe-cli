package cli

import "github.com/spf13/cobra"

func newEquipmentShowCmd() *cobra.Command {
	return newGenericShowCmd("equipment")
}

func init() {
	equipmentCmd.AddCommand(newEquipmentShowCmd())
}
