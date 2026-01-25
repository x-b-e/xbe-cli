package cli

import "github.com/spf13/cobra"

var doOneStepGpsVehiclesCmd = &cobra.Command{
	Use:     "one-step-gps-vehicles",
	Aliases: []string{"one-step-gps-vehicle"},
	Short:   "Manage One Step GPS vehicles",
	Long:    "Commands for updating One Step GPS vehicles.",
}

func init() {
	doCmd.AddCommand(doOneStepGpsVehiclesCmd)
}
