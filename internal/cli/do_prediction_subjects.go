package cli

import "github.com/spf13/cobra"

var doPredictionSubjectsCmd = &cobra.Command{
	Use:     "prediction-subjects",
	Aliases: []string{"prediction-subject"},
	Short:   "Manage prediction subjects",
	Long: `Manage prediction subjects on the XBE platform.

Prediction subjects capture questions, timelines, and outcomes for broker and
project predictions.

Commands:
  create    Create a prediction subject
  update    Update a prediction subject
  delete    Delete a prediction subject`,
	Example: `  # Create a prediction subject
  xbe do prediction-subjects create --name "Forecast" --parent-type brokers --parent-id 123 --status active

  # Update a prediction subject
  xbe do prediction-subjects update 123 --status complete --actual 125000

  # Delete a prediction subject
  xbe do prediction-subjects delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doPredictionSubjectsCmd)
}
