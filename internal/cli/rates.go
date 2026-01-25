package cli

import "github.com/spf13/cobra"

var ratesCmd = &cobra.Command{
	Use:     "rates",
	Aliases: []string{"rate"},
	Short:   "View rates",
	Long:    "Commands for viewing rates.",
}

func init() {
	viewCmd.AddCommand(ratesCmd)
}
