package cli

import "github.com/spf13/cobra"

var doShiftSummaryCmd = &cobra.Command{
	Use:   "shift-summary",
	Short: "Generate shift summaries",
	Long: `Generate shift summaries on the XBE platform.

Commands:
  create    Generate a shift summary`,
	Example: `  # Generate a shift summary grouped by driver
  xbe summarize shift-summary create --group-by driver --filter broker=123 --filter start_on=2025-01-01 --filter end_on=2025-01-31

  # Summary by trucker and date
  xbe summarize shift-summary create --group-by trucker,date --filter broker=123 --filter start_on=2025-01-01 --filter end_on=2025-01-31

  # Summary with specific metrics
  xbe summarize shift-summary create --group-by driver --filter broker=123 --filter start_on=2025-01-01 --filter end_on=2025-01-31 --metrics shift_count,hours_sum,tons_sum`,
}

func init() {
	summarizeCmd.AddCommand(doShiftSummaryCmd)
}
