package cli

import "github.com/spf13/cobra"

var samsaraVehiclesCmd = &cobra.Command{
	Use:     "samsara-vehicles",
	Aliases: []string{"samsara-vehicle"},
	Short:   "View samsara vehicles",
	Long:    "Commands for viewing samsara vehicles.",
}

func init() {
	viewCmd.AddCommand(samsaraVehiclesCmd)
}
