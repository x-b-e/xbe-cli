package cli

import "github.com/spf13/cobra"

var geotabVehiclesCmd = &cobra.Command{
	Use:     "geotab-vehicles",
	Aliases: []string{"geotab-vehicle"},
	Short:   "View geotab vehicles",
	Long:    "Commands for viewing geotab vehicles.",
}

func init() {
	viewCmd.AddCommand(geotabVehiclesCmd)
}
