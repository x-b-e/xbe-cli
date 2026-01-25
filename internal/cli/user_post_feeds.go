package cli

import "github.com/spf13/cobra"

var userPostFeedsCmd = &cobra.Command{
	Use:     "user-post-feeds",
	Aliases: []string{"user-post-feed"},
	Short:   "View user post feeds",
	Long:    "Commands for viewing user post feeds.",
}

func init() {
	viewCmd.AddCommand(userPostFeedsCmd)
}
