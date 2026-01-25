package cli

import "github.com/spf13/cobra"

var t3EquipmentshareVehiclesCmd = &cobra.Command{
	Use:     "t3-equipmentshare-vehicles",
	Aliases: []string{"t3-equipmentshare-vehicle"},
	Short:   "View T3 EquipmentShare vehicles",
	Long:    "Commands for viewing T3 EquipmentShare vehicles.",
}

func init() {
	viewCmd.AddCommand(t3EquipmentshareVehiclesCmd)
}
