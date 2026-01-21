package cli

import "github.com/spf13/cobra"

var doPTPTrailerSummaryCmd = &cobra.Command{
	Use:   "ptp-trailer-summary",
	Short: "Generate project transport plan trailer summaries",
	Long: `Generate project transport plan trailer summaries on the XBE platform.

Commands:
  create    Generate a project transport plan trailer summary`,
	Example: `  # Generate a PTP trailer summary grouped by trailer
  xbe summarize ptp-trailer-summary create --group-by trailer --filter broker=123 --filter ordered_date_min=2025-01-01 --filter ordered_date_max=2025-01-31

  # Summary by customer
  xbe summarize ptp-trailer-summary create --group-by customer --filter broker=123

  # Summary with prediction metrics
  xbe summarize ptp-trailer-summary create --group-by trailer --filter broker=123 --metrics count,trailer_assignment_prediction_correct_pct`,
}

func init() {
	summarizeCmd.AddCommand(doPTPTrailerSummaryCmd)
}
