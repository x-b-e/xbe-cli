package cli

import "github.com/spf13/cobra"

var teletracNavmanVehiclesCmd = &cobra.Command{
	Use:     "teletrac-navman-vehicles",
	Aliases: []string{"teletrac-navman-vehicle"},
	Short:   "Browse Teletrac Navman vehicles",
	Long: `Browse Teletrac Navman vehicle integrations.

Teletrac Navman vehicles represent vehicles imported from Teletrac Navman
integrations and show assignment status to tractors and trailers.

Commands:
  list    List Teletrac Navman vehicles with filtering
  show    Show Teletrac Navman vehicle details`,
	Example: `  # List Teletrac Navman vehicles
  xbe view teletrac-navman-vehicles list

  # Show Teletrac Navman vehicle details
  xbe view teletrac-navman-vehicles show 123`,
}

func init() {
	viewCmd.AddCommand(teletracNavmanVehiclesCmd)
}
