package cli

import "github.com/spf13/cobra"

var answersCmd = &cobra.Command{
	Use:     "answers",
	Aliases: []string{"answer"},
	Short:   "View answers",
	Long: `View answers generated for questions.

Answers contain the response content and reference the originating question.

Commands:
  list    List answers
  show    Show answer details`,
	Example: `  # List answers
  xbe view answers list

  # Filter by question
  xbe view answers list --question 123

  # Show answer details
  xbe view answers show 456`,
}

func init() {
	viewCmd.AddCommand(answersCmd)
}
