package cli

import "github.com/spf13/cobra"

var doVerizonRevealVehiclesCmd = &cobra.Command{
	Use:     "verizon-reveal-vehicles",
	Aliases: []string{"verizon-reveal-vehicle"},
	Short:   "Manage Verizon Reveal vehicles",
	Long:    "Commands for updating Verizon Reveal vehicles.",
}

func init() {
	doCmd.AddCommand(doVerizonRevealVehiclesCmd)
}
