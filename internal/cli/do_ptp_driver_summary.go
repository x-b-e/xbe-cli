package cli

import "github.com/spf13/cobra"

var doPTPDriverSummaryCmd = &cobra.Command{
	Use:   "ptp-driver-summary",
	Short: "Generate project transport plan driver summaries",
	Long: `Generate project transport plan driver summaries on the XBE platform.

Commands:
  create    Generate a project transport plan driver summary`,
	Example: `  # Generate a PTP driver summary grouped by driver
  xbe summarize ptp-driver-summary create --group-by driver --filter broker=123 --filter ordered_date_min=2025-01-01 --filter ordered_date_max=2025-01-31

  # Summary by customer
  xbe summarize ptp-driver-summary create --group-by customer --filter broker=123

  # Summary with prediction metrics
  xbe summarize ptp-driver-summary create --group-by driver --filter broker=123 --metrics count,confirmation_pct,driver_assignment_prediction_correct_pct`,
}

func init() {
	summarizeCmd.AddCommand(doPTPDriverSummaryCmd)
}
