package cli

import "github.com/spf13/cobra"

var doQuestionsCmd = &cobra.Command{
	Use:   "questions",
	Short: "Manage questions",
	Long: `Create, update, and delete questions.

Questions capture user-submitted prompts and can be triaged or assigned.

Commands:
  create    Create a question
  update    Update a question
  delete    Delete a question`,
	Example: `  # Create a question
  xbe do questions create --content "What are today's safety priorities?"

  # Update a question
  xbe do questions update 123 --is-triaged true

  # Delete a question
  xbe do questions delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doQuestionsCmd)
}
