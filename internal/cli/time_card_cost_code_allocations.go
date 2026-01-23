package cli

import "github.com/spf13/cobra"

var timeCardCostCodeAllocationsCmd = &cobra.Command{
	Use:     "time-card-cost-code-allocations",
	Aliases: []string{"time-card-cost-code-allocation"},
	Short:   "View time card cost code allocations",
	Long: `View time card cost code allocations.

Time card cost code allocations define how a time card is split across cost
codes using percentage-based allocations.

Commands:
  list    List time card cost code allocations
  show    Show time card cost code allocation details`,
	Example: `  # List time card cost code allocations
  xbe view time-card-cost-code-allocations list

  # Show a time card cost code allocation
  xbe view time-card-cost-code-allocations show 123`,
}

func init() {
	viewCmd.AddCommand(timeCardCostCodeAllocationsCmd)
}
