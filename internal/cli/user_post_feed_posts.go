package cli

import "github.com/spf13/cobra"

var userPostFeedPostsCmd = &cobra.Command{
	Use:     "user-post-feed-posts",
	Aliases: []string{"user-post-feed-post"},
	Short:   "Browse user post feed posts",
	Long: `Browse user post feed posts.

User post feed posts represent entries in a user's post feed.

Commands:
  list    List user post feed posts with filtering and pagination
  show    Show user post feed post details`,
	Example: `  # List user post feed posts
  xbe view user-post-feed-posts list

  # Show a user post feed post
  xbe view user-post-feed-posts show 123`,
}

func init() {
	viewCmd.AddCommand(userPostFeedPostsCmd)
}
