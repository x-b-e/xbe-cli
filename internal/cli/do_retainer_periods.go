package cli

import "github.com/spf13/cobra"

var doRetainerPeriodsCmd = &cobra.Command{
	Use:     "retainer-periods",
	Aliases: []string{"retainer-period"},
	Short:   "Manage retainer periods",
	Long:    "Commands for creating, updating, and deleting retainer periods.",
}

func init() {
	doCmd.AddCommand(doRetainerPeriodsCmd)
}
