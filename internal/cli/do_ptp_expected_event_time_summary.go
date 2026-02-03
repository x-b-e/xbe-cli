package cli

import "github.com/spf13/cobra"

var doPTPExpectedEventTimeSummaryCmd = &cobra.Command{
	Use:   "ptp-expected-event-time-summary",
	Short: "Generate project transport plan expected event time summaries",
	Long: `Generate project transport plan expected event time summaries on the XBE platform.

Commands:
  create    Generate a project transport plan expected event time summary`,
	Example: `  # Summary grouped by event type
  xbe summarize ptp-expected-event-time-summary create --group-by event_type --filter broker=123 --filter expected_created_date_min=2025-01-01 --filter expected_created_date_max=2025-01-31

  # Summary by lead time bin
  xbe summarize ptp-expected-event-time-summary create --group-by lead_time_bin --filter broker=123

  # Summary by time type
  xbe summarize ptp-expected-event-time-summary create --group-by time_type --filter broker=123 --metrics snapshot_count,error_minutes_avg`,
}

func init() {
	summarizeCmd.AddCommand(doPTPExpectedEventTimeSummaryCmd)
}
