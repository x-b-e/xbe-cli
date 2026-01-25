package cli

import "github.com/spf13/cobra"

var doTransportOrdersCmd = &cobra.Command{
	Use:     "transport-orders",
	Aliases: []string{"transport-order"},
	Short:   "Manage transport orders",
	Long:    `Create transport orders.`,
}

func init() {
	doCmd.AddCommand(doTransportOrdersCmd)
}
