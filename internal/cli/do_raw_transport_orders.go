package cli

import "github.com/spf13/cobra"

var doRawTransportOrdersCmd = &cobra.Command{
	Use:     "raw-transport-orders",
	Aliases: []string{"raw-transport-order"},
	Short:   "Manage raw transport orders",
	Long: `Create, update, and delete raw transport orders.

Raw transport orders store imported order payloads before they are normalized
into transport orders. Admin access is required for create/update/delete.`,
}

func init() {
	doCmd.AddCommand(doRawTransportOrdersCmd)
}
