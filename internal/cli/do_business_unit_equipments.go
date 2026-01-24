package cli

import "github.com/spf13/cobra"

var doBusinessUnitEquipmentsCmd = &cobra.Command{
	Use:   "business-unit-equipments",
	Short: "Manage business unit equipment links",
	Long:  "Commands for creating and deleting business unit equipment links.",
}

func init() {
	doCmd.AddCommand(doBusinessUnitEquipmentsCmd)
}
