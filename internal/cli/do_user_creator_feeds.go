package cli

import "github.com/spf13/cobra"

var doUserCreatorFeedsCmd = &cobra.Command{
	Use:   "user-creator-feeds",
	Short: "Manage user creator feeds",
	Long: `Create, update, and delete user creator feeds.

User creator feeds track which creators appear in a user's creator feed.

Commands:
  create  Create a user creator feed
  update  Update a user creator feed
  delete  Delete a user creator feed`,
	Example: `  # Create a user creator feed
  xbe do user-creator-feeds create

  # Update a user creator feed
  xbe do user-creator-feeds update 123 --user 456

  # Delete a user creator feed
  xbe do user-creator-feeds delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doUserCreatorFeedsCmd)
}
