package cli

import "github.com/spf13/cobra"

var predictionSubjectRecapsCmd = &cobra.Command{
	Use:   "prediction-subject-recaps",
	Short: "Browse prediction subject recaps",
	Long: `Browse prediction subject recaps.

Prediction subject recaps are markdown summaries generated for prediction
subjects to capture key outcomes and context.

Commands:
  list    List prediction subject recaps with filtering and pagination
  show    Show prediction subject recap details`,
	Example: `  # List prediction subject recaps
  xbe view prediction-subject-recaps list

  # Filter by prediction subject
  xbe view prediction-subject-recaps list --prediction-subject 123

  # Show a prediction subject recap
  xbe view prediction-subject-recaps show 456`,
}

func init() {
	viewCmd.AddCommand(predictionSubjectRecapsCmd)
}
