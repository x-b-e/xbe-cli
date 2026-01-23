package cli

import "github.com/spf13/cobra"

var commentsCmd = &cobra.Command{
	Use:     "comments",
	Aliases: []string{"comment"},
	Short:   "View comments",
	Long:    "Commands for viewing comments on various resources.",
}

func init() {
	viewCmd.AddCommand(commentsCmd)
}
