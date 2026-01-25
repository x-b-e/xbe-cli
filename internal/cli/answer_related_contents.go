package cli

import "github.com/spf13/cobra"

var answerRelatedContentsCmd = &cobra.Command{
	Use:   "answer-related-contents",
	Short: "Browse answer related contents",
	Long: `Browse related content items for answers.

Answer related contents link answers to other content (newsletters, glossary
terms, release notes, press releases, objectives, features, and questions)
with a similarity score.

Commands:
  list    List related content links with filtering
  show    Show details of a related content link`,
	Example: `  # List related content for answers
  xbe view answer-related-contents list

  # Filter by answer
  xbe view answer-related-contents list --answer 123

  # Filter by related content
  xbe view answer-related-contents list --related-content-type newsletters --related-content-id 456

  # Show a specific related content link
  xbe view answer-related-contents show 789`,
}

func init() {
	viewCmd.AddCommand(answerRelatedContentsCmd)
}
