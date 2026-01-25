package cli

import "github.com/spf13/cobra"

func newEquipmentLocationEstimatesShowCmd() *cobra.Command {
	return newGenericShowCmd("equipment-location-estimates")
}

func init() {
	equipmentLocationEstimatesCmd.AddCommand(newEquipmentLocationEstimatesShowCmd())
}
