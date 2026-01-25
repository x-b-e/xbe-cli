package cli

import "github.com/spf13/cobra"

var gaugeVehiclesCmd = &cobra.Command{
	Use:     "gauge-vehicles",
	Aliases: []string{"gauge-vehicle"},
	Short:   "View gauge vehicles",
	Long:    "Commands for viewing gauge vehicles.",
}

func init() {
	viewCmd.AddCommand(gaugeVehiclesCmd)
}
