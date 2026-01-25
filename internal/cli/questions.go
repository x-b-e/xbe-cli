package cli

import "github.com/spf13/cobra"

var questionsCmd = &cobra.Command{
	Use:     "questions",
	Aliases: []string{"question"},
	Short:   "View questions",
	Long: `Browse questions asked by users.

Questions capture user-submitted prompts that can be answered and triaged.

Commands:
  list    List questions
  show    Show question details`,
	Example: `  # List questions
  xbe view questions list

  # Show a question
  xbe view questions show 123`,
}

func init() {
	viewCmd.AddCommand(questionsCmd)
}
