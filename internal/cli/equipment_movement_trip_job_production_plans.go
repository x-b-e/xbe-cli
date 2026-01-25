package cli

import "github.com/spf13/cobra"

var equipmentMovementTripJobProductionPlansCmd = &cobra.Command{
	Use:     "equipment-movement-trip-job-production-plans",
	Aliases: []string{"equipment-movement-trip-job-production-plan"},
	Short:   "View equipment movement trip job production plans",
	Long: `Browse equipment movement trip job production plans.

Equipment movement trip job production plans link equipment movement trips
to job production plans.

Commands:
  list    List equipment movement trip job production plans
  show    Show equipment movement trip job production plan details`,
	Example: `  # List equipment movement trip job production plans
  xbe view equipment-movement-trip-job-production-plans list

  # Show a specific link
  xbe view equipment-movement-trip-job-production-plans show 123`,
}

func init() {
	viewCmd.AddCommand(equipmentMovementTripJobProductionPlansCmd)
}
