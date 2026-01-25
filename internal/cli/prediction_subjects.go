package cli

import "github.com/spf13/cobra"

var predictionSubjectsCmd = &cobra.Command{
	Use:     "prediction-subjects",
	Aliases: []string{"prediction-subject"},
	Short:   "Browse prediction subjects",
	Long: `Browse prediction subjects.

Prediction subjects capture questions, timelines, and outcomes for broker and
project predictions.

Commands:
  list    List prediction subjects
  show    Show prediction subject details`,
	Example: `  # List prediction subjects
  xbe view prediction-subjects list

  # Show a prediction subject
  xbe view prediction-subjects show 123`,
}

func init() {
	viewCmd.AddCommand(predictionSubjectsCmd)
}
