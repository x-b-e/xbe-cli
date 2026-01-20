package cli

import "github.com/spf13/cobra"

var doMaterialTransactionSummaryCmd = &cobra.Command{
	Use:   "material-transaction-summary",
	Short: "Generate material transaction summaries",
	Long: `Generate material transaction summaries on the XBE platform.

Commands:
  create    Generate a material transaction summary`,
	Example: `  # Generate a material transaction summary grouped by material site
  xbe summarize material-transaction-summary create --group-by material_site --filter broker=123 --filter date_min=2025-01-01 --filter date_max=2025-01-31

  # Summary by customer segment
  xbe summarize material-transaction-summary create --group-by customer_segment --filter broker=123

  # Summary by date with tons metrics
  xbe summarize material-transaction-summary create --group-by date --filter broker=123 --metrics tons_sum,tons_avg`,
}

func init() {
	summarizeCmd.AddCommand(doMaterialTransactionSummaryCmd)
}
