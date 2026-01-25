package cli

import "github.com/spf13/cobra"

var doTruckerShiftSetsCmd = &cobra.Command{
	Use:     "trucker-shift-sets",
	Aliases: []string{"trucker-shift-set"},
	Short:   "Manage trucker shift sets",
	Long:    "Commands for updating trucker shift sets (driver days).",
}

func init() {
	doCmd.AddCommand(doTruckerShiftSetsCmd)
}
