package cli

import "github.com/spf13/cobra"

var doGaugeVehiclesCmd = &cobra.Command{
	Use:     "gauge-vehicles",
	Aliases: []string{"gauge-vehicle"},
	Short:   "Manage gauge vehicles",
	Long:    "Commands for updating gauge vehicles.",
}

func init() {
	doCmd.AddCommand(doGaugeVehiclesCmd)
}
