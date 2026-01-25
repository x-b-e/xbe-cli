package cli

import "github.com/spf13/cobra"

var predictionSubjectGapPortionsCmd = &cobra.Command{
	Use:   "prediction-subject-gap-portions",
	Short: "Browse prediction subject gap portions",
	Long: `Browse prediction subject gap portions.

Prediction subject gap portions represent named amounts that explain a portion
of the gap on a prediction subject.

Commands:
  list    List prediction subject gap portions with filtering and pagination
  show    Show prediction subject gap portion details`,
	Example: `  # List prediction subject gap portions
  xbe view prediction-subject-gap-portions list

  # Filter by prediction subject gap
  xbe view prediction-subject-gap-portions list --prediction-subject-gap 123

  # Show a prediction subject gap portion
  xbe view prediction-subject-gap-portions show 456`,
}

func init() {
	viewCmd.AddCommand(predictionSubjectGapPortionsCmd)
}
