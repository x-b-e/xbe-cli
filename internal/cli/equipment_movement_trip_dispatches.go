package cli

import "github.com/spf13/cobra"

var equipmentMovementTripDispatchesCmd = &cobra.Command{
	Use:     "equipment-movement-trip-dispatches",
	Aliases: []string{"equipment-movement-trip-dispatch"},
	Short:   "Browse equipment movement trip dispatches",
	Long: `Browse equipment movement trip dispatches.

Equipment movement trip dispatches orchestrate the creation and assignment of
movement trips and capture status, inputs, and fulfillment results.

Commands:
  list    List trip dispatches with filtering and pagination
  show    Show full details of a trip dispatch`,
	Example: `  # List trip dispatches
  xbe view equipment-movement-trip-dispatches list

  # Filter by status
  xbe view equipment-movement-trip-dispatches list --status pending

  # Filter by equipment movement trip
  xbe view equipment-movement-trip-dispatches list --equipment-movement-trip 123

  # Show a trip dispatch
  xbe view equipment-movement-trip-dispatches show 456`,
}

func init() {
	viewCmd.AddCommand(equipmentMovementTripDispatchesCmd)
}
