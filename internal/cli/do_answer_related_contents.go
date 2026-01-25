package cli

import "github.com/spf13/cobra"

var doAnswerRelatedContentsCmd = &cobra.Command{
	Use:   "answer-related-contents",
	Short: "Manage answer related contents",
	Long: `Create, update, and delete answer related content links.

These links associate answers with related content items and are typically
admin-only operations.`,
}

func init() {
	doCmd.AddCommand(doAnswerRelatedContentsCmd)
}
