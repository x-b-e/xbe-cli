package cli

import "github.com/spf13/cobra"

var equipmentMovementTripCustomerCostAllocationsCmd = &cobra.Command{
	Use:   "equipment-movement-trip-customer-cost-allocations",
	Short: "View equipment movement trip customer cost allocations",
	Long: `View customer cost allocations for equipment movement trips.

These allocations define how trip costs are split across customers based on the
associated equipment movement requirements.

Commands:
  list    List equipment movement trip customer cost allocations
  show    Show equipment movement trip customer cost allocation details`,
	Example: `  # List cost allocations
  xbe view equipment-movement-trip-customer-cost-allocations list

  # Show a specific cost allocation
  xbe view equipment-movement-trip-customer-cost-allocations show 123`,
}

func init() {
	viewCmd.AddCommand(equipmentMovementTripCustomerCostAllocationsCmd)
}
