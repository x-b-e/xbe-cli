package cli

import "github.com/spf13/cobra"

var doCommentReactionsCmd = &cobra.Command{
	Use:     "comment-reactions",
	Aliases: []string{"comment-reaction"},
	Short:   "Manage comment reactions",
	Long:    "Commands for creating and deleting comment reactions.",
}

func init() {
	doCmd.AddCommand(doCommentReactionsCmd)
}
