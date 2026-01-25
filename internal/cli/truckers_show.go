package cli

import "github.com/spf13/cobra"

func newTruckersShowCmd() *cobra.Command {
	return newGenericShowCmd("truckers")
}

func init() {
	truckersCmd.AddCommand(newTruckersShowCmd())
}
