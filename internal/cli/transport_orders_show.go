package cli

import "github.com/spf13/cobra"

func newTransportOrdersShowCmd() *cobra.Command {
	return newGenericShowCmd("transport-orders")
}

func init() {
	transportOrdersCmd.AddCommand(newTransportOrdersShowCmd())
}
