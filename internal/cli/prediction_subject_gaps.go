package cli

import "github.com/spf13/cobra"

var predictionSubjectGapsCmd = &cobra.Command{
	Use:     "prediction-subject-gaps",
	Aliases: []string{"prediction-subject-gap"},
	Short:   "Browse prediction subject gaps",
	Long: `Browse prediction subject gaps.

Prediction subject gaps capture differences between primary and secondary
prediction values for a subject, along with approval status.

Commands:
  list    List prediction subject gaps
  show    Show prediction subject gap details`,
	Example: `  # List prediction subject gaps
  xbe view prediction-subject-gaps list

  # Filter by prediction subject
  xbe view prediction-subject-gaps list --prediction-subject 123

  # Show prediction subject gap details
  xbe view prediction-subject-gaps show 456`,
}

func init() {
	viewCmd.AddCommand(predictionSubjectGapsCmd)
}
