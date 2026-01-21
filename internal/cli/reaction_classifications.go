package cli

import "github.com/spf13/cobra"

var reactionClassificationsCmd = &cobra.Command{
	Use:   "reaction-classifications",
	Short: "View reaction classifications",
	Long: `View reaction classifications on the XBE platform.

Reaction classifications define the available emoji reactions that can be
used on posts, comments, and other content.

Note: Reaction classifications are read-only and cannot be created,
updated, or deleted through the API.

Commands:
  list    List reaction classifications`,
	Example: `  # List reaction classifications
  xbe view reaction-classifications list

  # Output as JSON
  xbe view reaction-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(reactionClassificationsCmd)
}
