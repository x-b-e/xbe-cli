package cli

import "github.com/spf13/cobra"

var gpsInsightVehiclesCmd = &cobra.Command{
	Use:     "gps-insight-vehicles",
	Aliases: []string{"gps-insight-vehicle"},
	Short:   "View GPS Insight vehicles",
	Long:    "Commands for viewing GPS Insight vehicles.",
}

func init() {
	viewCmd.AddCommand(gpsInsightVehiclesCmd)
}
