package cli

import "github.com/spf13/cobra"

var doEquipmentMovementTripJobProductionPlansCmd = &cobra.Command{
	Use:     "equipment-movement-trip-job-production-plans",
	Aliases: []string{"equipment-movement-trip-job-production-plan"},
	Short:   "Manage equipment movement trip job production plans",
	Long:    "Commands for creating and deleting equipment movement trip job production plans.",
}

func init() {
	doCmd.AddCommand(doEquipmentMovementTripJobProductionPlansCmd)
}
