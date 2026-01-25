package cli

import "github.com/spf13/cobra"

var equipmentRentalsCmd = &cobra.Command{
	Use:     "equipment-rentals",
	Aliases: []string{"equipment-rental"},
	Short:   "View equipment rentals",
	Long:    "Commands for viewing equipment rentals.",
}

func init() {
	viewCmd.AddCommand(equipmentRentalsCmd)
}
