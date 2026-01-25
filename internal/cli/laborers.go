package cli

import "github.com/spf13/cobra"

var laborersCmd = &cobra.Command{
	Use:     "laborers",
	Aliases: []string{"laborer"},
	Short:   "View laborers",
	Long:    "Commands for viewing laborers.",
}

func init() {
	viewCmd.AddCommand(laborersCmd)
}
