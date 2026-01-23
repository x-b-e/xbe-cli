package cli

import "github.com/spf13/cobra"

var doCrewRatesCmd = &cobra.Command{
	Use:     "crew-rates",
	Aliases: []string{"crew-rate"},
	Short:   "Manage crew rates",
	Long:    "Commands for creating, updating, and deleting crew rates.",
}

func init() {
	doCmd.AddCommand(doCrewRatesCmd)
}
