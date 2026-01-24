package cli

import "github.com/spf13/cobra"

var doPredictionsCmd = &cobra.Command{
	Use:     "predictions",
	Aliases: []string{"prediction"},
	Short:   "Manage predictions",
	Long: `Manage predictions.

Predictions capture probability distributions for a prediction subject. You can
create predictions, update attributes like status or distribution, or delete
predictions when permitted by policy.

Commands:
  create   Create a prediction
  update   Update a prediction
  delete   Delete a prediction`,
	Example: `  # Create a prediction
  xbe do predictions create \
    --prediction-subject 123 \
    --status draft \
    --probability-distribution '{"class_name":"TriangularDistribution","minimum":100,"mode":120,"maximum":140}'

  # Update prediction status
  xbe do predictions update 456 --status submitted

  # Delete a prediction
  xbe do predictions delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doPredictionsCmd)
}
