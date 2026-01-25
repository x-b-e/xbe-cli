package cli

import "github.com/spf13/cobra"

var doPredictionSubjectBidsCmd = &cobra.Command{
	Use:     "prediction-subject-bids",
	Aliases: []string{"prediction-subject-bid"},
	Short:   "Manage prediction subject bids",
	Long: `Manage prediction subject bids.

Prediction subject bids capture individual bidder amounts tied to a prediction
subject's lowest losing bid detail. You can create bids, update amounts, or
remove bids when permitted by policy.

Commands:
  create   Create a prediction subject bid
  update   Update a prediction subject bid
  delete   Delete a prediction subject bid`,
	Example: `  # Create a prediction subject bid
  xbe do prediction-subject-bids create \
    --bidder 123 \
    --lowest-losing-bid-prediction-subject-detail 456 \
    --amount 120000

  # Update a bid amount
  xbe do prediction-subject-bids update 789 --amount 125000

  # Delete a bid
  xbe do prediction-subject-bids delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doPredictionSubjectBidsCmd)
}
