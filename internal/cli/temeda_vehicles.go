package cli

import "github.com/spf13/cobra"

var temedaVehiclesCmd = &cobra.Command{
	Use:     "temeda-vehicles",
	Aliases: []string{"temeda-vehicle"},
	Short:   "View temeda vehicles",
	Long:    "Commands for viewing temeda vehicles.",
}

func init() {
	viewCmd.AddCommand(temedaVehiclesCmd)
}
