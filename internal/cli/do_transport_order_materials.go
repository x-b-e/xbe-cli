package cli

import "github.com/spf13/cobra"

var doTransportOrderMaterialsCmd = &cobra.Command{
	Use:     "transport-order-materials",
	Aliases: []string{"transport-order-material"},
	Short:   "Manage transport order materials",
	Long:    "Create, update, and delete transport order materials.",
}

func init() {
	doCmd.AddCommand(doTransportOrderMaterialsCmd)
}
