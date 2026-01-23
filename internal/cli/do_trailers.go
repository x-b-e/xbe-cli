package cli

import "github.com/spf13/cobra"

var doTrailersCmd = &cobra.Command{
	Use:     "trailers",
	Aliases: []string{"trailer"},
	Short:   "Manage trailers",
	Long:    "Commands for creating, updating, and deleting trailers.",
}

func init() {
	doCmd.AddCommand(doTrailersCmd)
}
