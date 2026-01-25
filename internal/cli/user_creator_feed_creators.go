package cli

import "github.com/spf13/cobra"

var userCreatorFeedCreatorsCmd = &cobra.Command{
	Use:   "user-creator-feed-creators",
	Short: "Browse user creator feed creators",
	Long: `Browse creators within user creator feeds.

User creator feed creators represent the ordered creators shown in a user's
creator feed.

Commands:
  list    List user creator feed creators with filtering and pagination
  show    Show details for a user creator feed creator`,
}

func init() {
	viewCmd.AddCommand(userCreatorFeedCreatorsCmd)
}
