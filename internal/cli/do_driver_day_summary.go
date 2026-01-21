package cli

import "github.com/spf13/cobra"

var doDriverDaySummaryCmd = &cobra.Command{
	Use:   "driver-day-summary",
	Short: "Generate driver day summaries",
	Long: `Generate driver day summaries on the XBE platform.

Commands:
  create    Generate a driver day summary`,
	Example: `  # Generate a driver day summary grouped by driver
  xbe summarize driver-day-summary create --group-by driver --filter broker=123 --filter start_on_min=2025-01-01 --filter start_on_max=2025-01-31

  # Summary by trucker
  xbe summarize driver-day-summary create --group-by trucker --filter broker=123 --filter start_on_min=2025-01-01 --filter start_on_max=2025-01-31

  # Summary with specific metrics
  xbe summarize driver-day-summary create --group-by driver --filter broker=123 --filter start_on_min=2025-01-01 --filter start_on_max=2025-01-31 --metrics driver_day_count,duration_hours_sum`,
}

func init() {
	summarizeCmd.AddCommand(doDriverDaySummaryCmd)
}
