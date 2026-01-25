package cli

import "github.com/spf13/cobra"

func newCommentsShowCmd() *cobra.Command {
	return newGenericShowCmd("comments")
}

func init() {
	commentsCmd.AddCommand(newCommentsShowCmd())
}
