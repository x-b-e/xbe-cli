package cli

import "github.com/spf13/cobra"

var doCycleTimeComparisonsCmd = &cobra.Command{
	Use:   "cycle-time-comparisons",
	Short: "Compare cycle times between two locations",
	Long: `Compare cycle times between two locations.

Cycle time comparisons estimate cycle durations between two coordinate pairs
using accepted material transactions within a proximity radius.

Commands:
  create    Create a cycle time comparison`,
	Example: `  # Compare cycle times between two points
  xbe do cycle-time-comparisons create \
    --coordinates-one '[37.7749,-122.4194]' \
    --coordinates-two '[37.8044,-122.2712]' \
    --proximity-meters 5000

  # Limit to a date range and sample size
  xbe do cycle-time-comparisons create \
    --coordinates-one '[37.7749,-122.4194]' \
    --coordinates-two '[37.8044,-122.2712]' \
    --proximity-meters 5000 \
    --transaction-at-min 2024-01-01T00:00:00Z \
    --transaction-at-max 2024-12-31T23:59:59Z \
    --cycle-count 250`,
}

func init() {
	doCmd.AddCommand(doCycleTimeComparisonsCmd)
}
