package cli

import "github.com/spf13/cobra"

var meetingsCmd = &cobra.Command{
	Use:     "meetings",
	Aliases: []string{"meeting"},
	Short:   "View meetings",
	Long:    "Commands for viewing meetings.",
}

func init() {
	viewCmd.AddCommand(meetingsCmd)
}
