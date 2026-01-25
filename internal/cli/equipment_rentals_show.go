package cli

import "github.com/spf13/cobra"

func newEquipmentRentalsShowCmd() *cobra.Command {
	return newGenericShowCmd("equipment-rentals")
}

func init() {
	equipmentRentalsCmd.AddCommand(newEquipmentRentalsShowCmd())
}
