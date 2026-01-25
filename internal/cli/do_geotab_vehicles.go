package cli

import "github.com/spf13/cobra"

var doGeotabVehiclesCmd = &cobra.Command{
	Use:     "geotab-vehicles",
	Aliases: []string{"geotab-vehicle"},
	Short:   "Manage geotab vehicles",
	Long:    "Commands for updating geotab vehicles.",
}

func init() {
	doCmd.AddCommand(doGeotabVehiclesCmd)
}
