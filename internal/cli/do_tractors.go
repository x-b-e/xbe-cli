package cli

import "github.com/spf13/cobra"

var doTractorsCmd = &cobra.Command{
	Use:     "tractors",
	Aliases: []string{"tractor"},
	Short:   "Manage tractors",
	Long:    "Commands for creating, updating, and deleting tractors.",
}

func init() {
	doCmd.AddCommand(doTractorsCmd)
}
