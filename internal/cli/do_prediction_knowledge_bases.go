package cli

import "github.com/spf13/cobra"

var doPredictionKnowledgeBasesCmd = &cobra.Command{
	Use:     "prediction-knowledge-bases",
	Aliases: []string{"prediction-knowledge-base"},
	Short:   "Manage prediction knowledge bases",
	Long: `Manage prediction knowledge bases on the XBE platform.

Prediction knowledge bases store broker-specific question repositories for
prediction subjects.

Commands:
  create    Create a prediction knowledge base`,
	Example: `  # Create a prediction knowledge base
  xbe do prediction-knowledge-bases create --broker 123`,
}

func init() {
	doCmd.AddCommand(doPredictionKnowledgeBasesCmd)
}
