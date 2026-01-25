package cli

import "github.com/spf13/cobra"

var doSamsaraVehiclesCmd = &cobra.Command{
	Use:     "samsara-vehicles",
	Aliases: []string{"samsara-vehicle"},
	Short:   "Manage samsara vehicles",
	Long:    "Commands for updating samsara vehicles.",
}

func init() {
	doCmd.AddCommand(doSamsaraVehiclesCmd)
}
