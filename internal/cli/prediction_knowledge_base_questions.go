package cli

import "github.com/spf13/cobra"

var predictionKnowledgeBaseQuestionsCmd = &cobra.Command{
	Use:   "prediction-knowledge-base-questions",
	Short: "View prediction knowledge base questions",
	Long: `Browse prediction knowledge base questions.

Prediction knowledge base questions capture the prompts used to generate
knowledge base answers for prediction subjects.

Commands:
  list    List prediction knowledge base questions
  show    Show prediction knowledge base question details`,
	Example: `  # List knowledge base questions
  xbe view prediction-knowledge-base-questions list

  # Show a knowledge base question
  xbe view prediction-knowledge-base-questions show 123`,
}

func init() {
	viewCmd.AddCommand(predictionKnowledgeBaseQuestionsCmd)
}
