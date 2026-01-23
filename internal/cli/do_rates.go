package cli

import "github.com/spf13/cobra"

var doRatesCmd = &cobra.Command{
	Use:     "rates",
	Aliases: []string{"rate"},
	Short:   "Manage rates",
	Long:    "Commands for creating, updating, and deleting rates.",
}

func init() {
	doCmd.AddCommand(doRatesCmd)
}
