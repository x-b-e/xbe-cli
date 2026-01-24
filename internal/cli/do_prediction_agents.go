package cli

import "github.com/spf13/cobra"

var doPredictionAgentsCmd = &cobra.Command{
	Use:   "prediction-agents",
	Short: "Manage prediction agents",
	Long: `Create, update, and delete prediction agents.

Prediction agents generate crowd forecasts for prediction subjects and can
create or update associated predictions.

Commands:
  create    Create a prediction agent
  update    Update a prediction agent
  delete    Delete a prediction agent`,
	Example: `  # Create a prediction agent
  xbe do prediction-agents create --prediction-subject 123

  # Update custom instructions
  xbe do prediction-agents update 456 --custom-instructions "Focus on recent data"

  # Delete a prediction agent
  xbe do prediction-agents delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doPredictionAgentsCmd)
}
