package cli

import "github.com/spf13/cobra"

var doEquipmentRentalsCmd = &cobra.Command{
	Use:     "equipment-rentals",
	Aliases: []string{"equipment-rental"},
	Short:   "Manage equipment rentals",
	Long:    "Commands for creating, updating, and deleting equipment rentals.",
}

func init() {
	doCmd.AddCommand(doEquipmentRentalsCmd)
}
