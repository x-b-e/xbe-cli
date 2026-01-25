package cli

import "github.com/spf13/cobra"

var doUserPostFeedPostsCmd = &cobra.Command{
	Use:     "user-post-feed-posts",
	Aliases: []string{"user-post-feed-post"},
	Short:   "Manage user post feed posts",
	Long: `Manage user post feed posts.

Commands:
  update  Update a user post feed post`,
	Example: `  # Update a user post feed post
  xbe do user-post-feed-posts update 123 --is-bookmarked=true`,
}

func init() {
	doCmd.AddCommand(doUserPostFeedPostsCmd)
}
