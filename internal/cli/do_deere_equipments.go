package cli

import "github.com/spf13/cobra"

var doDeereEquipmentsCmd = &cobra.Command{
	Use:     "deere-equipments",
	Aliases: []string{"deere-equipment"},
	Short:   "Manage Deere equipment",
	Long: `Manage Deere equipment assignments on the XBE platform.

Deere equipment records are created from integrations. They cannot be created or
deleted via the API, but they can be updated for assignment and metadata fixes.`,
}

func init() {
	doCmd.AddCommand(doDeereEquipmentsCmd)
}
