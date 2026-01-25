package cli

import "github.com/spf13/cobra"

var doKeepTruckinVehiclesCmd = &cobra.Command{
	Use:     "keep-truckin-vehicles",
	Aliases: []string{"keep-truckin-vehicle"},
	Short:   "Manage KeepTruckin vehicles",
	Long: `Manage KeepTruckin vehicle assignments on the XBE platform.

KeepTruckin vehicles are created from integrations. They cannot be created or
removed via the API, but trailer and tractor assignments can be updated.`,
}

func init() {
	doCmd.AddCommand(doKeepTruckinVehiclesCmd)
}
