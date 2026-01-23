package cli

import "github.com/spf13/cobra"

var doEquipmentMovementTripCustomerCostAllocationsCmd = &cobra.Command{
	Use:   "equipment-movement-trip-customer-cost-allocations",
	Short: "Manage equipment movement trip customer cost allocations",
	Long: `Create, update, and delete customer cost allocations for equipment movement trips.

Each equipment movement trip may have a single customer cost allocation that
splits costs across customers associated with the trip's requirements.

Commands:
  create    Create a customer cost allocation
  update    Update a customer cost allocation
  delete    Delete a customer cost allocation`,
}

func init() {
	doCmd.AddCommand(doEquipmentMovementTripCustomerCostAllocationsCmd)
}
