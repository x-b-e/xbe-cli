package cli

import "github.com/spf13/cobra"

var doGpsInsightVehiclesCmd = &cobra.Command{
	Use:     "gps-insight-vehicles",
	Aliases: []string{"gps-insight-vehicle"},
	Short:   "Manage GPS Insight vehicles",
	Long:    "Commands for updating GPS Insight vehicles.",
}

func init() {
	doCmd.AddCommand(doGpsInsightVehiclesCmd)
}
