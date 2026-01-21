package cli

import "github.com/spf13/cobra"

var doPTPEventTimeSummaryCmd = &cobra.Command{
	Use:   "ptp-event-time-summary",
	Short: "Generate project transport plan event time summaries",
	Long: `Generate project transport plan event time summaries on the XBE platform.

Commands:
  create    Generate a project transport plan event time summary`,
	Example: `  # Generate a PTP event time summary grouped by event type
  xbe summarize ptp-event-time-summary create --group-by event_type --filter broker=123 --filter event_date_local_min=2025-01-01 --filter event_date_local_max=2025-01-31

  # Summary by location
  xbe summarize ptp-event-time-summary create --group-by location --filter broker=123

  # Summary with duration metrics
  xbe summarize ptp-event-time-summary create --group-by event_type --filter broker=123 --metrics event_count,duration_minutes_avg,duration_minutes_p90`,
}

func init() {
	summarizeCmd.AddCommand(doPTPEventTimeSummaryCmd)
}
