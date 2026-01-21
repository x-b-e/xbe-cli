package cli

import "github.com/spf13/cobra"

var doPTPEventSummaryCmd = &cobra.Command{
	Use:   "ptp-event-summary",
	Short: "Generate project transport plan event summaries",
	Long: `Generate project transport plan event summaries on the XBE platform.

Commands:
  create    Generate a project transport plan event summary`,
	Example: `  # Generate a PTP event summary grouped by event type
  xbe summarize ptp-event-summary create --group-by event_type --filter broker=123 --filter created_date_min=2025-01-01 --filter created_date_max=2025-01-31

  # Summary by broker
  xbe summarize ptp-event-summary create --group-by broker --filter broker=123

  # Summary with prediction metrics
  xbe summarize ptp-event-summary create --group-by event_type --filter broker=123 --metrics count,location_assignment_prediction_correct_pct`,
}

func init() {
	summarizeCmd.AddCommand(doPTPEventSummaryCmd)
}
