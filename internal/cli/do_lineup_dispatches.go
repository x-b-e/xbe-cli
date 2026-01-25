package cli

import "github.com/spf13/cobra"

var doLineupDispatchesCmd = &cobra.Command{
	Use:     "lineup-dispatches",
	Aliases: []string{"lineup-dispatch"},
	Short:   "Manage lineup dispatches",
	Long:    "Commands for creating, updating, and deleting lineup dispatches.",
}

func init() {
	doCmd.AddCommand(doLineupDispatchesCmd)
}
