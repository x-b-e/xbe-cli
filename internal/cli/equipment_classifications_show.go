package cli

import "github.com/spf13/cobra"

func newEquipmentClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("equipment-classifications")
}

func init() {
	equipmentClassificationsCmd.AddCommand(newEquipmentClassificationsShowCmd())
}
