package cli

import "github.com/spf13/cobra"

var parkingSitesCmd = &cobra.Command{
	Use:     "parking-sites",
	Aliases: []string{"parking-site"},
	Short:   "View parking sites",
	Long:    "Commands for viewing parking sites.",
}

func init() {
	viewCmd.AddCommand(parkingSitesCmd)
}
