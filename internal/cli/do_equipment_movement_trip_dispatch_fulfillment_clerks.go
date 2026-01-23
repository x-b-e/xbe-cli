package cli

import "github.com/spf13/cobra"

var doEquipmentMovementTripDispatchFulfillmentClerksCmd = &cobra.Command{
	Use:     "equipment-movement-trip-dispatch-fulfillment-clerks",
	Aliases: []string{"equipment-movement-trip-dispatch-fulfillment-clerk"},
	Short:   "Fulfill equipment movement trip dispatches",
	Long:    "Commands for running the fulfillment clerk on equipment movement trip dispatches.",
}

func init() {
	doCmd.AddCommand(doEquipmentMovementTripDispatchFulfillmentClerksCmd)
}
