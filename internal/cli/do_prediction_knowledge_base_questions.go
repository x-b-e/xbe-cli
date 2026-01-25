package cli

import "github.com/spf13/cobra"

var doPredictionKnowledgeBaseQuestionsCmd = &cobra.Command{
	Use:   "prediction-knowledge-base-questions",
	Short: "Manage prediction knowledge base questions",
	Long: `Create, update, and delete prediction knowledge base questions.

Knowledge base questions capture the prompts used to generate answers for
prediction subjects.

Commands:
  create    Create a prediction knowledge base question
  update    Update a prediction knowledge base question
  delete    Delete a prediction knowledge base question`,
	Example: `  # Create a knowledge base question
  xbe do prediction-knowledge-base-questions create --prediction-knowledge-base 123 --title "What are the key risks?"

  # Update a knowledge base question
  xbe do prediction-knowledge-base-questions update 456 --status resolved

  # Delete a knowledge base question
  xbe do prediction-knowledge-base-questions delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doPredictionKnowledgeBaseQuestionsCmd)
}
