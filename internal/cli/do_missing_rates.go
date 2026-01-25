package cli

import "github.com/spf13/cobra"

var doMissingRatesCmd = &cobra.Command{
	Use:     "missing-rates",
	Aliases: []string{"missing-rate"},
	Short:   "Manage missing rates",
	Long:    "Commands for creating missing rates.",
}

func init() {
	doCmd.AddCommand(doMissingRatesCmd)
}
