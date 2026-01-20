package cli

import "github.com/spf13/cobra"

var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Aggregate data for analysis (pivot tables, totals, statistics)",
	Long: `Aggregate large datasets for analysis.

Summary commands work like pivot tables, grouping and aggregating data to produce
totals, averages, and other statistics. Use these when you need to analyze trends
or compare metrics across dimensions.

Resources:
  lane-summary                     Aggregate hauling/cycle data by origin, destination, etc.
  material-transaction-summary     Aggregate material transactions by site, customer, date, etc.`,
	Example: `  # Summarize hauling data by origin and destination
  xbe summarize lane-summary create --group-by origin,destination --filter broker=123

  # Summarize material transactions by site for a date range
  xbe summarize material-transaction-summary create \
    --group-by material_site \
    --filter broker=123 --filter date_min=2025-01-01

  # Get totals without grouping
  xbe summarize lane-summary create --group-by "" --filter broker=123`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}
