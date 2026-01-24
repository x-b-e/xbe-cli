package cli

import "github.com/spf13/cobra"

var doLowestLosingBidPredictionSubjectDetailsCmd = &cobra.Command{
	Use:   "lowest-losing-bid-prediction-subject-details",
	Short: "Manage lowest losing bid prediction subject details",
	Long: `Create, update, and delete lowest losing bid prediction subject details.

These records track bid amounts and estimate data for lowest losing bid prediction subjects.

Commands:
  create    Create a detail record
  update    Update a detail record
  delete    Delete a detail record`,
}

func init() {
	doCmd.AddCommand(doLowestLosingBidPredictionSubjectDetailsCmd)
}
