package cli

import "github.com/spf13/cobra"

var doTripsCmd = &cobra.Command{
	Use:     "trips",
	Aliases: []string{"trip"},
	Short:   "Manage trips",
	Long:    "Commands for creating, updating, and deleting trips.",
}

func init() {
	doCmd.AddCommand(doTripsCmd)
}
