package cli

import "github.com/spf13/cobra"

var doPredictionSubjectGapPortionsCmd = &cobra.Command{
	Use:   "prediction-subject-gap-portions",
	Short: "Manage prediction subject gap portions",
	Long: `Create, update, and delete prediction subject gap portions.

Prediction subject gap portions explain components of a prediction subject gap.

Commands:
  create    Create a prediction subject gap portion
  update    Update a prediction subject gap portion
  delete    Delete a prediction subject gap portion`,
	Example: `  # Create a prediction subject gap portion
  xbe do prediction-subject-gap-portions create --prediction-subject-gap 123 --name "Labor" --amount 42 --status draft

  # Update a prediction subject gap portion
  xbe do prediction-subject-gap-portions update 456 --status approved

  # Delete a prediction subject gap portion
  xbe do prediction-subject-gap-portions delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doPredictionSubjectGapPortionsCmd)
}
