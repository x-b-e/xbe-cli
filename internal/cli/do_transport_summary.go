package cli

import "github.com/spf13/cobra"

var doTransportSummaryCmd = &cobra.Command{
	Use:   "transport-summary",
	Short: "Generate transport summaries",
	Long: `Generate transport summaries on the XBE platform.

Commands:
  create    Generate a transport summary`,
	Example: `  # Generate a transport order summary
  xbe summarize transport-summary create --entity-type transport_order --filter broker=123 --filter start_date=2025-01-01 --filter end_date=2025-01-31

  # Generate a transport plan summary
  xbe summarize transport-summary create --entity-type transport_plan --filter broker=123

  # Generate a driver assignment summary
  xbe summarize transport-summary create --entity-type driver_assignment --filter broker=123

  # Generate a live loads summary
  xbe summarize transport-summary create --entity-type live_loads --filter broker=123`,
}

func init() {
	summarizeCmd.AddCommand(doTransportSummaryCmd)
}
