package cli

import "github.com/spf13/cobra"

var tennaVehiclesCmd = &cobra.Command{
	Use:     "tenna-vehicles",
	Aliases: []string{"tenna-vehicle"},
	Short:   "Browse Tenna vehicles",
	Long: `Browse Tenna vehicle integrations.

Tenna vehicles represent vehicles imported from Tenna integrations and show
assignment status to tractors, trailers, and equipment.

Commands:
  list    List Tenna vehicles with filtering
  show    Show Tenna vehicle details`,
	Example: `  # List Tenna vehicles
  xbe view tenna-vehicles list

  # Show Tenna vehicle details
  xbe view tenna-vehicles show 123`,
}

func init() {
	viewCmd.AddCommand(tennaVehiclesCmd)
}
