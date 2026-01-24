package cli

import "github.com/spf13/cobra"

var predictionsCmd = &cobra.Command{
	Use:     "predictions",
	Aliases: []string{"prediction"},
	Short:   "Browse predictions",
	Long: `Browse predictions.

Predictions capture probability distributions for a prediction subject, along
with status and scoring metadata.

Commands:
  list    List predictions
  show    Show prediction details`,
	Example: `  # List predictions
  xbe view predictions list

  # Filter by prediction subject
  xbe view predictions list --prediction-subject 123

  # Show prediction details
  xbe view predictions show 456`,
}

func init() {
	viewCmd.AddCommand(predictionsCmd)
}
