package cli

import "github.com/spf13/cobra"

var oneStepGpsVehiclesCmd = &cobra.Command{
	Use:     "one-step-gps-vehicles",
	Aliases: []string{"one-step-gps-vehicle"},
	Short:   "View One Step GPS vehicles",
	Long:    "Commands for viewing One Step GPS vehicles.",
}

func init() {
	viewCmd.AddCommand(oneStepGpsVehiclesCmd)
}
