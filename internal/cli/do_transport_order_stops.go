package cli

import "github.com/spf13/cobra"

var doTransportOrderStopsCmd = &cobra.Command{
	Use:     "transport-order-stops",
	Aliases: []string{"transport-order-stop"},
	Short:   "Manage transport order stops",
	Long:    "Commands for creating, updating, and deleting transport order stops.",
}

func init() {
	doCmd.AddCommand(doTransportOrderStopsCmd)
}
