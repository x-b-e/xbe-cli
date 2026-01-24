package cli

import "github.com/spf13/cobra"

var doMarketingMetricsCmd = &cobra.Command{
	Use:   "marketing-metrics",
	Short: "Refresh marketing metrics",
	Long: `Refresh marketing metrics.

Marketing metrics are aggregated counters cached on the server. The create
command refreshes the cached snapshot and returns the current values.

Commands:
  create    Refresh marketing metrics`,
	Example: `  # Refresh marketing metrics
  xbe do marketing-metrics create

  # Output as JSON
  xbe do marketing-metrics create --json`,
}

func init() {
	doCmd.AddCommand(doMarketingMetricsCmd)
}
