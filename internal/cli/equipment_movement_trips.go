package cli

import "github.com/spf13/cobra"

var equipmentMovementTripsCmd = &cobra.Command{
	Use:     "equipment-movement-trips",
	Aliases: []string{"equipment-movement-trip"},
	Short:   "View equipment movement trips",
	Long:    "Commands for viewing equipment movement trips.",
}

func init() {
	viewCmd.AddCommand(equipmentMovementTripsCmd)
}
