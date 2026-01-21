package cli

import "github.com/spf13/cobra"

var doPTPSummaryCmd = &cobra.Command{
	Use:   "ptp-summary",
	Short: "Generate project transport plan summaries",
	Long: `Generate project transport plan summaries on the XBE platform.

Commands:
  create    Generate a project transport plan summary`,
	Example: `  # Generate a PTP summary grouped by broker
  xbe summarize ptp-summary create --group-by broker --filter broker=123 --filter created_date_min=2025-01-01 --filter created_date_max=2025-01-31

  # Summary by strategy set
  xbe summarize ptp-summary create --group-by strategy_set --filter broker=123

  # Summary with prediction metrics
  xbe summarize ptp-summary create --group-by broker --filter broker=123 --metrics count,strategy_set_prediction_correct_pct`,
}

func init() {
	summarizeCmd.AddCommand(doPTPSummaryCmd)
}
