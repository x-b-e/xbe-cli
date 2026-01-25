package cli

import "github.com/spf13/cobra"

var tractorsCmd = &cobra.Command{
	Use:     "tractors",
	Aliases: []string{"tractor"},
	Short:   "View tractors",
	Long:    "Commands for viewing tractors.",
}

func init() {
	viewCmd.AddCommand(tractorsCmd)
}
