package cli

import "github.com/spf13/cobra"

var doTennaVehiclesCmd = &cobra.Command{
	Use:     "tenna-vehicles",
	Aliases: []string{"tenna-vehicle"},
	Short:   "Manage Tenna vehicles",
	Long: `Manage Tenna vehicle assignments on the XBE platform.

Tenna vehicles are created from integrations. They cannot be created or
removed via the API, but trailer, tractor, and equipment assignments can be updated.`,
}

func init() {
	doCmd.AddCommand(doTennaVehiclesCmd)
}
