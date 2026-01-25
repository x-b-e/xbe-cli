package cli

import "github.com/spf13/cobra"

var predictionSubjectBidsCmd = &cobra.Command{
	Use:     "prediction-subject-bids",
	Aliases: []string{"prediction-subject-bid"},
	Short:   "Browse prediction subject bids",
	Long: `Browse prediction subject bids.

Prediction subject bids capture individual bidder amounts tied to a prediction
subject's lowest losing bid detail.

Commands:
  list    List prediction subject bids
  show    Show prediction subject bid details`,
	Example: `  # List prediction subject bids
  xbe view prediction-subject-bids list

  # Filter by bidder
  xbe view prediction-subject-bids list --bidder 123

  # Show prediction subject bid details
  xbe view prediction-subject-bids show 456`,
}

func init() {
	viewCmd.AddCommand(predictionSubjectBidsCmd)
}
