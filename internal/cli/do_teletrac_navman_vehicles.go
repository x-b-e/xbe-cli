package cli

import "github.com/spf13/cobra"

var doTeletracNavmanVehiclesCmd = &cobra.Command{
	Use:     "teletrac-navman-vehicles",
	Aliases: []string{"teletrac-navman-vehicle"},
	Short:   "Manage Teletrac Navman vehicles",
	Long: `Manage Teletrac Navman vehicle assignments on the XBE platform.

Teletrac Navman vehicles are created from integrations. They cannot be created or
removed via the API, but trailer and tractor assignments can be updated.`,
}

func init() {
	doCmd.AddCommand(doTeletracNavmanVehiclesCmd)
}
