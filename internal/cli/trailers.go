package cli

import "github.com/spf13/cobra"

var trailersCmd = &cobra.Command{
	Use:     "trailers",
	Aliases: []string{"trailer"},
	Short:   "View trailers",
	Long:    "Commands for viewing trailers.",
}

func init() {
	viewCmd.AddCommand(trailersCmd)
}
