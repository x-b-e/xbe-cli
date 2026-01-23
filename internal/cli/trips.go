package cli

import "github.com/spf13/cobra"

var tripsCmd = &cobra.Command{
	Use:     "trips",
	Aliases: []string{"trip"},
	Short:   "View trips",
	Long:    "Commands for viewing trips.",
}

func init() {
	viewCmd.AddCommand(tripsCmd)
}
