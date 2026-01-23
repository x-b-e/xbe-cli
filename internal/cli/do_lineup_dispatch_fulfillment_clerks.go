package cli

import "github.com/spf13/cobra"

var doLineupDispatchFulfillmentClerksCmd = &cobra.Command{
	Use:     "lineup-dispatch-fulfillment-clerks",
	Aliases: []string{"lineup-dispatch-fulfillment-clerk"},
	Short:   "Fulfill lineup dispatches",
	Long:    "Commands for running the fulfillment clerk on lineup dispatches.",
}

func init() {
	doCmd.AddCommand(doLineupDispatchFulfillmentClerksCmd)
}
