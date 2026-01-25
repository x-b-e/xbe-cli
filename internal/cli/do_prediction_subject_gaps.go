package cli

import "github.com/spf13/cobra"

var doPredictionSubjectGapsCmd = &cobra.Command{
	Use:     "prediction-subject-gaps",
	Aliases: []string{"prediction-subject-gap"},
	Short:   "Manage prediction subject gaps",
	Long: `Manage prediction subject gaps.

Prediction subject gaps capture differences between primary and secondary
amounts for a prediction subject, and may be approved or updated when
permitted by policy.

Commands:
  create   Create a prediction subject gap
  update   Update a prediction subject gap
  delete   Delete a prediction subject gap`,
	Example: `  # Create a prediction subject gap
  xbe do prediction-subject-gaps create \
    --prediction-subject 123 \
    --gap-type actual_vs_consensus

  # Approve a gap
  xbe do prediction-subject-gaps update 456 --status approved

  # Delete a gap
  xbe do prediction-subject-gaps delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doPredictionSubjectGapsCmd)
}
