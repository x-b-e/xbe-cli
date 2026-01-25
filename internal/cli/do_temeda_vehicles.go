package cli

import "github.com/spf13/cobra"

var doTemedaVehiclesCmd = &cobra.Command{
	Use:     "temeda-vehicles",
	Aliases: []string{"temeda-vehicle"},
	Short:   "Manage temeda vehicles",
	Long:    "Commands for updating temeda vehicles.",
}

func init() {
	doCmd.AddCommand(doTemedaVehiclesCmd)
}
