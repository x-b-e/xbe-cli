package cli

import "github.com/spf13/cobra"

var crewRatesCmd = &cobra.Command{
	Use:     "crew-rates",
	Aliases: []string{"crew-rate"},
	Short:   "View crew rates",
	Long:    "Commands for viewing crew rates.",
}

func init() {
	viewCmd.AddCommand(crewRatesCmd)
}
