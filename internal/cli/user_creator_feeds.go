package cli

import "github.com/spf13/cobra"

var userCreatorFeedsCmd = &cobra.Command{
	Use:     "user-creator-feeds",
	Aliases: []string{"user-creator-feed"},
	Short:   "View user creator feeds",
	Long:    "Commands for viewing user creator feeds.",
}

func init() {
	viewCmd.AddCommand(userCreatorFeedsCmd)
}
