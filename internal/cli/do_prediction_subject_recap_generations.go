package cli

import "github.com/spf13/cobra"

var doPredictionSubjectRecapGenerationsCmd = &cobra.Command{
	Use:   "prediction-subject-recap-generations",
	Short: "Generate prediction subject recaps",
	Long: `Generate prediction subject recaps.

Prediction subject recap generations schedule recap creation for a prediction
subject.

Commands:
  create    Generate a prediction subject recap`,
	Example: `  # Generate a recap for a prediction subject
  xbe do prediction-subject-recap-generations create --prediction-subject 123

  # JSON output
  xbe do prediction-subject-recap-generations create --prediction-subject 123 --json`,
}

func init() {
	doCmd.AddCommand(doPredictionSubjectRecapGenerationsCmd)
}
