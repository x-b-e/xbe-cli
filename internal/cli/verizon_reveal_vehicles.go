package cli

import "github.com/spf13/cobra"

var verizonRevealVehiclesCmd = &cobra.Command{
	Use:     "verizon-reveal-vehicles",
	Aliases: []string{"verizon-reveal-vehicle"},
	Short:   "View Verizon Reveal vehicles",
	Long:    "Commands for viewing Verizon Reveal vehicles.",
}

func init() {
	viewCmd.AddCommand(verizonRevealVehiclesCmd)
}
