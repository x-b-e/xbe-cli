package cli

import "github.com/spf13/cobra"

func newReactionClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("reaction-classifications")
}

func init() {
	reactionClassificationsCmd.AddCommand(newReactionClassificationsShowCmd())
}
