package cli

import "github.com/spf13/cobra"

var lowestLosingBidPredictionSubjectDetailsCmd = &cobra.Command{
	Use:   "lowest-losing-bid-prediction-subject-details",
	Short: "Browse lowest losing bid prediction subject details",
	Long: `Browse lowest losing bid prediction subject details.

These records capture bid amounts and estimates for lowest losing bid prediction subjects.

Commands:
  list    List details with filtering and pagination
  show    Show full detail for a record`,
	Example: `  # List lowest losing bid prediction subject details
  xbe view lowest-losing-bid-prediction-subject-details list

  # Show details
  xbe view lowest-losing-bid-prediction-subject-details show 123`,
}

func init() {
	viewCmd.AddCommand(lowestLosingBidPredictionSubjectDetailsCmd)
}
