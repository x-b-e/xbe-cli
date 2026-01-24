package cli

import "github.com/spf13/cobra"

var marketingMetricsCmd = &cobra.Command{
	Use:     "marketing-metrics",
	Aliases: []string{"marketing-metric"},
	Short:   "Browse marketing metrics",
	Long: `Browse marketing metrics.

Marketing metrics are aggregated counters cached on the server. Use the list
command to refresh and view the latest snapshot.

Commands:
  list    Fetch and display the latest marketing metrics`,
	Example: `  # Fetch the latest marketing metrics
  xbe view marketing-metrics list

  # Output as JSON
  xbe view marketing-metrics list --json`,
}

func init() {
	viewCmd.AddCommand(marketingMetricsCmd)
}
