package cli

import "github.com/spf13/cobra"

var doParkingSitesCmd = &cobra.Command{
	Use:     "parking-sites",
	Aliases: []string{"parking-site"},
	Short:   "Manage parking sites",
	Long:    "Commands for creating, updating, and deleting parking sites.",
}

func init() {
	doCmd.AddCommand(doParkingSitesCmd)
}
