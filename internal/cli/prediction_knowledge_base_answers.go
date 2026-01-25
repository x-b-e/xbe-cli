package cli

import "github.com/spf13/cobra"

var predictionKnowledgeBaseAnswersCmd = &cobra.Command{
	Use:     "prediction-knowledge-base-answers",
	Aliases: []string{"prediction-knowledge-base-answer"},
	Short:   "Browse prediction knowledge base answers",
	Long: `Browse prediction knowledge base answers.

Prediction knowledge base answers are generated responses tied to prediction
knowledge base questions.

Commands:
  list    List prediction knowledge base answers
  show    Show prediction knowledge base answer details`,
	Example: `  # List prediction knowledge base answers
  xbe view prediction-knowledge-base-answers list

  # Show a prediction knowledge base answer
  xbe view prediction-knowledge-base-answers show 123`,
}

func init() {
	viewCmd.AddCommand(predictionKnowledgeBaseAnswersCmd)
}
