package cli

import "github.com/spf13/cobra"

var businessUnitEquipmentsCmd = &cobra.Command{
	Use:     "business-unit-equipments",
	Aliases: []string{"business-unit-equipment"},
	Short:   "View business unit equipment links",
	Long:    "Commands for viewing business unit equipment links.",
}

func init() {
	viewCmd.AddCommand(businessUnitEquipmentsCmd)
}
