package cli

import "github.com/spf13/cobra"

var predictionKnowledgeBasesCmd = &cobra.Command{
	Use:     "prediction-knowledge-bases",
	Aliases: []string{"prediction-knowledge-base"},
	Short:   "Browse prediction knowledge bases",
	Long: `Browse prediction knowledge bases.

Prediction knowledge bases store broker-specific question repositories for
prediction subjects.

Commands:
  list    List prediction knowledge bases
  show    Show prediction knowledge base details`,
	Example: `  # List prediction knowledge bases
  xbe view prediction-knowledge-bases list

  # Show a prediction knowledge base
  xbe view prediction-knowledge-bases show 123`,
}

func init() {
	viewCmd.AddCommand(predictionKnowledgeBasesCmd)
}
