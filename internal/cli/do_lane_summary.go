package cli

import "github.com/spf13/cobra"

var doLaneSummaryCmd = &cobra.Command{
	Use:   "lane-summary",
	Short: "Generate lane (cycle) summaries",
	Long: `Generate lane (cycle) summaries on the XBE platform.

Commands:
  create    Generate a lane summary (cycle summary)`,
	Example: `  # Generate a lane summary grouped by origin and destination
  xbe summarize lane-summary create --group-by origin,destination --filter broker=123 --filter transaction_at_min=2025-01-17T00:00:00Z --filter transaction_at_max=2025-01-17T23:59:59Z

  # Total summary (no group-by)
  xbe summarize lane-summary create --group-by "" --filter broker=123 --filter date_min=2025-01-01 --filter date_max=2025-01-31

  # Include driver day trip lead minutes and driver movement segment durations
  xbe summarize lane-summary create --filter broker=123 --use-driver-day-trip-lead-minutes --beta-driver-movement-segment-durations`,
}

func init() {
	summarizeCmd.AddCommand(doLaneSummaryCmd)
}
