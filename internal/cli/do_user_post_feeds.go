package cli

import "github.com/spf13/cobra"

var doUserPostFeedsCmd = &cobra.Command{
	Use:   "user-post-feeds",
	Short: "Manage user post feeds",
	Long: `Create, update, and delete user post feeds.

User post feeds track posts shown in a user's feed and vector indexing settings.

Commands:
  create  Create a user post feed
  update  Update a user post feed
  delete  Delete a user post feed`,
	Example: `  # Create a user post feed
  xbe do user-post-feeds create

  # Update vector indexing for a user post feed
  xbe do user-post-feeds update 123 --enable-vector-indexing true

  # Delete a user post feed
  xbe do user-post-feeds delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doUserPostFeedsCmd)
}
