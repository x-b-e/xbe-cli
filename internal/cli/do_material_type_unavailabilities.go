package cli

import "github.com/spf13/cobra"

var doMaterialTypeUnavailabilitiesCmd = &cobra.Command{
	Use:     "material-type-unavailabilities",
	Aliases: []string{"material-type-unavailability"},
	Short:   "Manage material type unavailabilities",
	Long:    "Commands for creating, updating, and deleting material type unavailabilities.",
}

func init() {
	doCmd.AddCommand(doMaterialTypeUnavailabilitiesCmd)
}
