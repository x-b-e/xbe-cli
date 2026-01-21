package cli

import "github.com/spf13/cobra"

var doTransportOrderEfficiencySummaryCmd = &cobra.Command{
	Use:   "transport-order-efficiency-summary",
	Short: "Generate transport order efficiency summaries",
	Long: `Generate transport order efficiency summaries on the XBE platform.

Commands:
  create    Generate a transport order efficiency summary`,
	Example: `  # Generate a transport order efficiency summary grouped by customer
  xbe summarize transport-order-efficiency-summary create --group-by customer --filter broker=123 --filter ordered_date_min=2025-01-01 --filter ordered_date_max=2025-01-31

  # Summary by driver
  xbe summarize transport-order-efficiency-summary create --group-by driver --filter broker=123

  # Summary with routing metrics
  xbe summarize transport-order-efficiency-summary create --group-by customer --filter broker=123 --metrics transport_order_count,routed_miles_sum,deviated_miles_sum`,
}

func init() {
	summarizeCmd.AddCommand(doTransportOrderEfficiencySummaryCmd)
}
