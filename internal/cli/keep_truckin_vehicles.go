package cli

import "github.com/spf13/cobra"

var keepTruckinVehiclesCmd = &cobra.Command{
	Use:     "keep-truckin-vehicles",
	Aliases: []string{"keep-truckin-vehicle"},
	Short:   "Browse KeepTruckin vehicles",
	Long: `Browse KeepTruckin vehicle integrations.

KeepTruckin vehicles represent vehicles imported from KeepTruckin
integrations and show assignment status to tractors and trailers.

Commands:
  list    List KeepTruckin vehicles with filtering
  show    Show KeepTruckin vehicle details`,
	Example: `  # List KeepTruckin vehicles
  xbe view keep-truckin-vehicles list

  # Show KeepTruckin vehicle details
  xbe view keep-truckin-vehicles show 123`,
}

func init() {
	viewCmd.AddCommand(keepTruckinVehiclesCmd)
}
