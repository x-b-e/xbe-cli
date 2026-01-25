package cli

import "github.com/spf13/cobra"

var doT3EquipmentshareVehiclesCmd = &cobra.Command{
	Use:     "t3-equipmentshare-vehicles",
	Aliases: []string{"t3-equipmentshare-vehicle"},
	Short:   "Manage T3 EquipmentShare vehicles",
	Long:    "Commands for updating T3 EquipmentShare vehicles.",
}

func init() {
	doCmd.AddCommand(doT3EquipmentshareVehiclesCmd)
}
